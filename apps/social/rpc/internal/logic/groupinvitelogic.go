package logic

import (
	"context"

	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLogic {
	return &GroupInviteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupInviteLogic) GroupInvite(in *social.GroupInviteReq) (*social.GroupInviteResp, error) {
	// 1. Check if inviter is in the group
	_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "inviter not in group")
	}

	// 2. Loop through friends to invite
	for _, friendId := range in.FriendIds {
		// Check if already a member
		_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, friendId, in.GroupId)
		if err == nil {
			continue // Already a member, skip
		}

		// Add as member
		// TODO: In a real app, this might create a request or notification first.
		// For now, we directly add them as members (like scanning QR code or direct invite in some apps).
		// Or we could create a GroupRequests record with JoinSource = InviteGroupJoinSource.

		// Direct add approach:
		groupMember := &socialmodels.GroupMembers{
			GroupId:     in.GroupId,
			UserId:      friendId,
			RoleLevel:   int(constants.AtLargeGroupRoleLevel),
			JoinTime:    time.Now(),
			JoinSource:  int(constants.InviteGroupJoinSource),
			InviterUid:  in.UserId,
			OperatorUid: in.UserId,
		}
		_, err = l.svcCtx.GroupMembersModel.Insert(l.ctx, groupMember)
		if err != nil {
			l.Logger.Errorf("Failed to invite user %s: %v", friendId, err)
			// Continue with others
		}
	}

	return &social.GroupInviteResp{}, nil
}
