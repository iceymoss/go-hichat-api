package group

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupInvitationCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInvitationCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationCreateLogic {
	return &GroupInvitationCreateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GroupInvitationCreateLogic) GroupInvitationCreate(req *types.GroupInvitationCreateReq) (*types.GroupInvitationCreateResp, error) {
	uid, err := apiActor(l.ctx)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.Social.GroupInvitationCreate(l.ctx, &social.GroupInvitationCreateReq{
		ActorUid: uid, GroupId: req.GroupId, InviteeUid: req.InviteeUid, Message: req.Message,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupInvitationCreateResp{Invitation: invitationType(res.Invitation)}, nil
}

func apiActor(ctx context.Context) (string, error) {
	uid := ctxdata.GetUId(ctx)
	if id, err := strconv.ParseUint(uid, 10, 64); err != nil || id == 0 {
		return "", status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	return uid, nil
}

func invitationType(invitation *social.GroupInvitation) types.GroupInvitation {
	if invitation == nil {
		return types.GroupInvitation{}
	}
	return types.GroupInvitation{
		Id: strconv.FormatUint(invitation.Id, 10), GroupId: invitation.GroupId, InviterUid: invitation.InviterUid, InviteeUid: invitation.InviteeUid,
		InviterRoleSnapshot: invitation.InviterRoleSnapshot, Message: invitation.Message, Status: invitation.Status,
		RejectReason: invitation.RejectReason, CreatedAt: invitation.CreatedAt, HandledAt: invitation.HandledAt, ExpiresAt: invitation.ExpiresAt,
		ReadState: invitation.ReadState, Actionable: invitation.Actionable,
	}
}
