package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUpdateLogic {
	return &GroupUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupUpdateLogic) GroupUpdate(in *social.GroupUpdateReq) (*social.GroupUpdateResp, error) {
	// 1. Check permission
	operator, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "operator not found")
	}

	if operator.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "user %s has no permission to update group", in.UserId)
	}

	// 2. Find Group
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found")
	}

	// 3. Update fields
	if in.Name != "" {
		group.Name = in.Name
	}
	if in.Icon != "" {
		group.Icon = in.Icon
	}
	if in.Notification != "" {
		group.Notification = in.Notification
		// Set notification uid to current operator
		uidInt, _ := strconv.Atoi(in.UserId)
		group.NotificationUid = uidInt
	}

	// 4. Save
	err = l.svcCtx.GroupsModel.Update(l.ctx, group)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group err %v", err)
	}

	return &social.GroupUpdateResp{}, nil
}
