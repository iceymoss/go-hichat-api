package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadLikesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadLikesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadLikesLogic {
	return &GetUnreadLikesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取未读点赞通知
func (l *GetUnreadLikesLogic) GetUnreadLikes(in *trend.GetUnreadLikesRequest) (*trend.GetUnreadLikesResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.GetUnreadLikesResponse{}, nil
}
