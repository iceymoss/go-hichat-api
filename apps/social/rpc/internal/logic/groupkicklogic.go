package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupKickLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupKickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupKickLogic {
	return &GroupKickLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupKickLogic) GroupKick(in *social.GroupKickReq) (*social.GroupKickResp, error) {
	// 1. Check if operator is admin/creator
	operator, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "operator not found")
	}

	if operator.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "user %s has no permission to kick", in.UserId)
	}

	// 2. Loop through members to kick
	for _, memberId := range in.MemberIds {
		// Find member
		member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, memberId, in.GroupId)
		if err == nil {
			// Cannot kick someone with higher or equal role
			if member.RoleLevel >= operator.RoleLevel {
				continue // Skip
			}
			l.svcCtx.GroupMembersModel.Delete(l.ctx, member.Id)
		}
	}

	return &social.GroupKickResp{}, nil
}
