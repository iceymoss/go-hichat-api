package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GroupQuitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupQuitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupQuitLogic {
	return &GroupQuitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupQuitLogic) GroupQuit(in *social.GroupQuitReq) (*social.GroupQuitResp, error) {
	// 1. Check if group exists
	_, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found %v", in.GroupId)
	}

	// 2. Find member record
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "member not found %v", err)
	}

	// 微信式约束：群主不能直接退群，必须先转让群主或解散群
	if member.RoleLevel == int(constants.CreatorGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("群主不能直接退群，请先转让群主或解散群"), "owner cannot quit")
	}

	// 3. Delete member + 同事务写 outbox（退群事件），提交后 best-effort 同步关系缓存
	err = emitRelationChange(l.ctx, l.svcCtx, constants.RelationEventGroupMemberRemoved, in.GroupId,
		&mq.RelationChangeTransfer{GroupId: in.GroupId, UserId: in.UserId, OperatorId: in.UserId},
		func(tx *gorm.DB) error {
			return tx.Table("group_members").Where("id = ?", member.Id).Delete(nil).Error
		})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "delete member err %v", err)
	}

	return &social.GroupQuitResp{}, nil
}
