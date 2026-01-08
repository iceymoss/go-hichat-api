package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 好友申请已读
func NewFriendPutInReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInReadLogic {
	return &FriendPutInReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInReadLogic) FriendPutInRead(req *types.FriendPutInReadReq) (resp *types.FriendPutInReadResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.FriendPutInRead(l.ctx, &socialclient.FriendPutInReadReq{
		UserId:      uid,
		FriendReqId: req.FriendReqId,
	})
	return &types.FriendPutInReadResp{}, err
}
