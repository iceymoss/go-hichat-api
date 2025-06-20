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

// 标记点赞为已读
func (l *MarkLikesReadLogic) MarkLikesRead(in *trend.MarkLikesReadRequest) (*trend.MarkLikesReadResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.MarkLikesReadResponse{}, nil
}
