package like

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikedUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取点赞用户列表
func NewGetLikedUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikedUsersLogic {
	return &GetLikedUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLikedUsersLogic) GetLikedUsers(req *types.GetLikedUsersRequest) (resp *types.GetLikedUsersResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
