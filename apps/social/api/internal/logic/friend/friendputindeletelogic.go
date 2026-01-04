package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除好友申请
func NewFriendPutInDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInDeleteLogic {
	return &FriendPutInDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInDeleteLogic) FriendPutInDelete(req *types.FriendPutInDeleteReq) (resp *types.FriendPutInDeleteResp, err error) {
	// todo: add your logic here and delete this line

	return
}
