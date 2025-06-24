package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikedUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLikedUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikedUsersLogic {
	return &GetLikedUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLikedUsers 获取点赞用户列表 (微信"查看全部点赞")
func (l *GetLikedUsersLogic) GetLikedUsers(in *trend.GetLikedUsersRequest) (*trend.GetLikedUsersResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.GetLikedUsersResponse{}, nil
}
