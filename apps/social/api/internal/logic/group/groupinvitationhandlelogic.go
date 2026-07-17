package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupInvitationHandleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInvitationHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationHandleLogic {
	return &GroupInvitationHandleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GroupInvitationHandleLogic) GroupInvitationHandle(req *types.GroupInvitationHandleReq) (*types.GroupInvitationHandleResp, error) {
	uid, err := apiActor(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "invitation id must be positive")
	}
	if req.Result != 1 && req.Result != 2 {
		return nil, status.Error(codes.InvalidArgument, "result must be 1 or 2")
	}
	res, err := l.svcCtx.Social.GroupInvitationHandle(l.ctx, &social.GroupInvitationHandleReq{
		Id: req.Id, ActorUid: uid, Result: req.Result, HandleMsg: req.HandleMsg,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupInvitationHandleResp{
		InvitationId: res.InvitationId, Status: res.Status, JoinState: res.JoinState,
		GroupRequestId: res.GroupRequestId, Idempotent: res.Idempotent, GroupId: res.GroupId,
	}, nil
}
