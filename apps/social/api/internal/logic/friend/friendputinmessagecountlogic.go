package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInMessageCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendPutInMessageCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInMessageCountLogic {
	return &FriendPutInMessageCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInMessageCountLogic) FriendPutInMessageCount(req *types.FriendPutInMessageCountReq) (resp *types.FriendPutInMessageCountResp, err error) {
	curUid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}

	// 调用 RPC 获取消息数量
	rpcResp, err := l.svcCtx.Social.FriendPutInMessageCount(l.ctx, &social.FriendPutInMessageCountReq{
		UserId: curUid, ActorUid: curUid,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.FriendPutInMessageCountResp{
		Count:  rpcResp.Count,
		Apply:  rpcResp.Apply,
		Result: rpcResp.Result,
	}
	return
}
