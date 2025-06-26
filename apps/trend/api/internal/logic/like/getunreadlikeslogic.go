package like

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadLikesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取未读点赞通知
func NewGetUnreadLikesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadLikesLogic {
	return &GetUnreadLikesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUnreadLikesLogic) GetUnreadLikes(req *types.GetUnreadLikesRequest) (resp *types.GetUnreadLikesResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
