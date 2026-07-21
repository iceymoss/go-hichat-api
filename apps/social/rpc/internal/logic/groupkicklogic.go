package logic

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
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
		if err == gorm.ErrRecordNotFound {
			return nil, errors.Wrapf(xerr.NewMsg("group not found"), "group %s not found", in.GroupId)
		}
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
			// 删除成员 + 同事务写 outbox（踢人事件），提交后 best-effort 同步关系缓存
			if emitErr := emitRelationChange(l.ctx, l.svcCtx, constants.RelationEventGroupMemberRemoved, in.GroupId,
				&mq.RelationChangeTransfer{GroupId: in.GroupId, UserId: memberId, OperatorId: in.UserId},
				func(tx *gorm.DB) error {
					return tx.Table("group_members").Where("id = ?", member.Id).Delete(nil).Error
				}); emitErr != nil {
				return nil, errors.Wrapf(xerr.NewDBErr(), "kick member err %v", emitErr)
			}
			// 公共通知：被移出群者收到通知
			emitCommonNotify(l.ctx, l.svcCtx, &mq.CommonNotify{
				NotifyType: NotifyGroupRemoved,
				ReceiverId: memberId,
				ActorId:    in.UserId,
				BizId:      fmt.Sprintf("group.removed:%s:%s:%d", in.GroupId, memberId, time.Now().Unix()),
				GroupId:    in.GroupId,
			})
		}
	}

	return &social.GroupKickResp{}, nil
}
