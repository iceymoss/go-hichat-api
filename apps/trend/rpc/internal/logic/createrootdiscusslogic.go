package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type CreateRootDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRootDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRootDiscussLogic {
	return &CreateRootDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateRootDiscuss 发表评论
func (l *CreateRootDiscussLogic) CreateRootDiscuss(in *trend.CreateDiscussReq) (*trend.CreateDiscussResp, error) {
	if in.TrendId == 0 {
		return nil, errors.New("动态ID不能为空")
	}
	if in.Content == "" || len(in.Content) > 500 {
		return nil, errors.New("评论内容无效或过长")
	}

	if in.Replyer == 0 {
		return nil, errors.New("用户未登录或会话已过期")
	}

	// 获取动态详情
	trendDetail, err := l.svcCtx.Trend.FindOne(l.ctx, in.TrendId)
	if err != nil || trendDetail == nil || trendDetail.State != 1 {
		return nil, errors.New("动态不存在或已被删除")
	}

	// 是否可评论
	if trendDetail.OpenReply != 1 {
		return nil, errors.New("此动态已关闭评论")
	}

	// 敏感词过滤
	pass, sensitiveWord, err := checkTrendContent(in.Content)
	if err != nil {
		return nil, err
	}

	if !pass {
		return nil, errors.New(fmt.Sprintf("动态标题或内容包含敏感词:%s", sensitiveWord))
	}

	// 创建评论实体
	now := time.Now()
	discuss := models.TrendDiscuss{
		Trendid:    int64(in.TrendId),
		Father:     0, // 一级评论 一级评论没有父亲评论
		Rootid:     0, // 一级评论 rootid=0 顶级评论
		Replyer:    in.Replyer,
		Userid:     in.UserId,
		Level:      1, // 一级评论
		Content:    in.Content,
		Idlist:     serializeAtUserIds(in.AtUserIds), // JSON 序列化 @用户列表
		State:      1,                                // 正常状态
		Read:       0,                                // 未读
		Path:       "/",
		Createtime: now,
		Updatetime: now,
	}

	// 存储评论
	id, err := l.svcCtx.TrendDiscuss.Insert(l.ctx, &discuss)
	if err != nil {
		zLog.Error("CreateRootDiscuss.Insert: create discuss", zap.Error(err))
		return nil, errors.New("评论失败")
	}

	// 更新动态评论数
	err = l.svcCtx.Trend.IncAgreeOrReply(l.ctx, in.TrendId, 1, 1)
	if err != nil {
		zLog.Error("CreateRootDiscuss.IncAgreeOrReply: inc reply failed", zap.Error(err))
		// 不中断处理，业务上可以接受丢失
	}

	// 不是评论自己，需要通知状态作者，被评论者
	if trendDetail.Userid != in.Replyer {
		go pushMessageWs(l.ctx, discuss)
	}

	return &trend.CreateDiscussResp{
		Discuss: &trend.Discuss{
			Id: id,
		},
	}, nil
}

// serializeAtUserIds 序列化 @用户列表
func serializeAtUserIds(userIds []uint64) string {
	if len(userIds) == 0 {
		return ""
	}
	data, _ := json.Marshal(userIds)
	return string(data)
}
