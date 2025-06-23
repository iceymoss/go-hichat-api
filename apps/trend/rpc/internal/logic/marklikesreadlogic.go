package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkLikesReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkLikesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkLikesReadLogic {
	return &MarkLikesReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkLikesRead 标记点赞为已读
func (l *MarkLikesReadLogic) MarkLikesRead(in *trend.MarkLikesReadRequest) (*trend.MarkLikesReadResponse, error) {
	err := l.svcCtx.TrendAgree.MarkRead(l.ctx, in.LikeIds)
	if err != nil {
		return nil, err
	}

	return &trend.MarkLikesReadResponse{}, nil
}
