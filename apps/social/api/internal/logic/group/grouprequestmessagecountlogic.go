package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupRequestMessageCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 群申请与邀请未读数量
func NewGroupRequestMessageCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupRequestMessageCountLogic {
	return &GroupRequestMessageCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupRequestMessageCountLogic) GroupRequestMessageCount(req *types.GroupRequestMessageCountReq) (resp *types.GroupRequestMessageCountResp, err error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}
	rpcResp, err := l.svcCtx.Social.GroupRequestMessageCount(l.ctx, &socialclient.GroupRequestMessageCountReq{UserId: uid, ActorUid: uid})
	if err != nil {
		return nil, err
	}
	return &types.GroupRequestMessageCountResp{Count: rpcResp.Count, Apply: rpcResp.Apply, Result: rpcResp.Result, Invite: rpcResp.Invite}, nil
}
