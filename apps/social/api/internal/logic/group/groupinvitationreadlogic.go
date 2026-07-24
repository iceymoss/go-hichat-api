package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInvitationReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记收到的群邀请已读
func NewGroupInvitationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationReadLogic {
	return &GroupInvitationReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupInvitationReadLogic) GroupInvitationRead(req *types.GroupInvitationReadReq) (resp *types.GroupInvitationReadResp, err error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}
	invitationIDs, err := parseGroupIDs(req.InvitationIds, "invitation id")
	if err != nil {
		return nil, err
	}
	rpcResp, err := l.svcCtx.Social.GroupInvitationRead(l.ctx, &socialclient.GroupInvitationReadReq{ActorUid: uid, InvitationIds: invitationIDs})
	if err != nil {
		return nil, err
	}
	return &types.GroupInvitationReadResp{Count: rpcResp.Count, Apply: rpcResp.Apply, Result: rpcResp.Result, Invite: rpcResp.Invite}, nil
}
