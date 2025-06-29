package logic

import (
	"context"
	"errors"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetChildDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChildDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildDiscussesLogic {
	return &GetChildDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetChildDiscusses 获取子评论（指定父评论下的所有评论）
func (l *GetChildDiscussesLogic) GetChildDiscusses(in *trend.GetChildDiscussesReq) (*trend.DiscussesListResp, error) {
	// 验证请求参数
	if in.Father == 0 {
		return nil, errors.New("一级评论ID不能为空")
	}

	// 获取一级评论的信息
	parentDiscuss, err := l.svcCtx.TrendDiscuss.FindOne(l.ctx, in.Father)
	if err != nil || parentDiscuss == nil || parentDiscuss.State != 1 {
		return nil, errors.New("一级评论不存在或已被删除")
	}

	// 获取总评论数
	var list []*trend.Discuss
	allCount := parentDiscuss.DiscussCount
	if allCount > 0 {
		childDiscusses, err := l.svcCtx.TrendDiscuss.FindChildrenByFather(l.ctx, in.Father, uint64(in.Pagination.LastId), 30)
		if err != nil {
			zLog.Error("GetChildDiscusses.FindChildrenByRootId: 查询二级评论失败", zap.Any("rootId", in.RootId), zap.Error(err))
			return nil, errors.New("获取评论失败")
		}

		// 获取id
		list = make([]*trend.Discuss, 0, len(childDiscusses))
		for _, v := range childDiscusses {
			list = append(list, convertToReply(v))
		}
	}
	last := &trend.Discuss{}
	if len(list) > 0 {
		last = list[len(list)-1]
	}

	return &trend.DiscussesListResp{
		Discusses: list,
		Total:     uint64(allCount),
		Pagination: &trend.Pagination{
			LastId:   int32(last.Id),
			LastTime: last.CreateTime,
		},
	}, nil
}
