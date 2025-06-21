package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type CreateChildDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChildDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChildDiscussLogic {
	return &CreateChildDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateChildDiscuss 发表多级评论（二级、三级等）
func (l *CreateChildDiscussLogic) CreateChildDiscuss(in *trend.CreateDiscussReq) (*trend.CreateDiscussResp, error) {
	// 1. 验证请求参数
	if in.Father == 0 {
		return nil, errors.New("父评论ID不能为空")
	}
	if in.Content == "" || len(in.Content) > 500 {
		return nil, errors.New("评论内容无效或过长")
	}

	// 获取用户ID
	if in.Replyer == 0 {
		return nil, errors.New("用户未登录或会话已过期")
	}

	// 获取父评论详情
	parentDiscuss, err := l.svcCtx.TrendDiscuss.FindOne(l.ctx, in.Father)
	if err != nil || parentDiscuss == nil {
		return nil, errors.New("父评论不存在或已被删除")
	}

	// 获取动态详情
	trendDetail, err := l.svcCtx.Trend.FindOne(l.ctx, uint64(parentDiscuss.Trendid))
	if err != nil || trendDetail == nil || trendDetail.State != 1 {
		return nil, errors.New("关联动态不存在或已被删除")
	}

	// 检查动态是否允许评论
	if trendDetail.OpenReply != 1 {
		return nil, errors.New("此动态已关闭评论功能")
	}

	// 检查父评论的状态
	if parentDiscuss.State != 1 {
		return nil, errors.New("该评论已被删除，无法回复")
	}

	// 敏感词过滤
	pass, sensitiveWord, err := checkTrendContent(in.Content)
	if err != nil {
		return nil, err
	}

	if !pass {
		return nil, errors.New(fmt.Sprintf("动态标题或内容包含敏感词:%s", sensitiveWord))
	}

	// 确定评论层级,父亲评论级别+1
	newLevel := parentDiscuss.Level + 1
	if newLevel > 20 {
		return nil, errors.New("评论层级过深，无法继续回复")
	}

	// 获取根评论id, 非二级评论直接继承
	rootId := parentDiscuss.Rootid
	if newLevel == 2 {
		//是二级评论,获取一级评论id，作为根id
		rootId = parentDiscuss.Id
	}

	// 创建评论实体
	now := time.Now()
	discuss := models.TrendDiscuss{
		Trendid:      parentDiscuss.Trendid,                                      // 动态ID从父评论获取
		Father:       int64(in.Father),                                           // 父评论ID
		Replyer:      in.Replyer,                                                 // 评论者ID
		Rootid:       rootId,                                                     // 根评论id
		Userid:       in.UserId,                                                  // 被回复用户ID（通常为父评论的作者）
		Level:        newLevel,                                                   // 层级
		Content:      in.Content,                                                 // 评论内容
		Idlist:       serializeAtUserIds(in.AtUserIds),                           // JSON 序列化 @用户列表
		AgreeCount:   0,                                                          // 点赞数
		DiscussCount: 0,                                                          // 子评论数
		State:        1,                                                          // 状态（1=正常）
		Read:         1,                                                          // 未读
		Path:         fmt.Sprintf("%s%d/", parentDiscuss.Path, parentDiscuss.Id), //评论路径，通过path前缀可以查询到当前评论的所有子评论
		Createtime:   now,                                                        // 创建时间
		Updatetime:   now,                                                        // 更新时间
	}

	var id uint64
	tx := transaction.NewManager()
	txErr := tx.Execute(l.ctx, nil, func(ctx context.Context) error {
		// 存储评论
		id, err = l.svcCtx.TrendDiscuss.Insert(l.ctx, &discuss)
		if err != nil {
			zLog.Error("CreateChildDiscuss.Insert: create child discuss failed", zap.Any("faterId", discuss.Father), zap.Any("trendId", discuss.Trendid), zap.Error(err))
			return errors.New("评论失败")
		}

		// 更新父评论的子评论计数
		parentDiscuss.DiscussCount += 1
		if err = l.svcCtx.TrendDiscuss.IncAgreeOrDiscuss(l.ctx, uint64(discuss.Father), 1, 1); err != nil {
			l.Logger.Error("更新父评论计数失败", logx.Field("error", err))
			zLog.Error("CreateChildDiscuss.IncAgreeOrDiscuss: 更新父评论计数失败", zap.Any("fatherId", discuss.Father), zap.Error(err))
			return err
		}

		if newLevel > 2 {
			// 更新根评论的评论数
			if err = l.svcCtx.TrendDiscuss.IncAgreeOrDiscuss(l.ctx, uint64(rootId), 1, 1); err != nil {
				l.Logger.Error("更新根评论计数失败", logx.Field("error", err))
				zLog.Error("CreateChildDiscuss.IncAgreeOrDiscuss: 更新根评论计数失败", zap.Any("root", discuss.Father), zap.Error(err))
				return err
			}
		}

		// 更新动态总评论数
		err = l.svcCtx.Trend.IncAgreeOrReply(l.ctx, in.TrendId, 1, 1)
		if err != nil {
			zLog.Error("CreateChildDiscuss.IncAgreeOrReply: inc reply failed", zap.Error(err))
			return err
		}

		return nil
	})
	if txErr != nil {
		zLog.Error("CreateChildDiscuss.Execute: 执行事务失败", zap.Any("root", discuss.Father), zap.Error(err))
		return nil, txErr
	}

	// 13. 发送通知（如果被回复用户不是自己）需要通知状态作者，被评论者
	if parentDiscuss.Replyer != in.Replyer {
		go pushMessageWs(l.ctx, discuss)
	}

	return &trend.CreateDiscussResp{
		Discuss: &trend.Discuss{
			Id:         uint64(id),
			TrendId:    uint64(parentDiscuss.Trendid),
			Father:     in.Father,
			RootId:     discuss.Rootid,
			Replyer:    in.Replyer,
			UserId:     in.UserId,
			Level:      uint32(newLevel),
			Content:    in.Content,
			AtUserIds:  in.AtUserIds,
			CreateTime: now.Unix(),
		},
	}, nil
}

func pushMessageWs(ctx context.Context, data any) error {
	// todo: 推送消息评论消息通知用户
	return nil
}
