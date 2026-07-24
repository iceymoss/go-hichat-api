package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm/clause"
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
	groupID, err := parsePositiveID(in.GroupId, "group id")
	if err != nil {
		return nil, err
	}
	operatorID, err := parsePositiveID(in.UserId, "operator uid")
	if err != nil {
		return nil, err
	}
	memberIDs := make([]uint64, 0, len(in.MemberIds))
	seen := make(map[uint64]struct{}, len(in.MemberIds))
	for _, memberID := range in.MemberIds {
		parsed, parseErr := parsePositiveID(memberID, "member uid")
		if parseErr != nil {
			return nil, parseErr
		}
		if _, duplicate := seen[parsed]; duplicate {
			continue
		}
		seen[parsed] = struct{}{}
		memberIDs = append(memberIDs, parsed)
	}

	type removedMember struct {
		userID  uint64
		event   *mq.RelationChangeTransfer
		version uint64
	}
	var removed []removedMember
	err = transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
		attemptRemoved := make([]removedMember, 0, len(memberIDs))
		query := tx
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var group objects.Group
		if err := query.First(&group, groupID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "group not found")
			}
			return err
		}
		if group.Status != nil && *group.Status != 0 {
			return status.Error(codes.FailedPrecondition, "group is unavailable")
		}
		lockedIDs := append([]uint64{operatorID}, memberIDs...)
		members, err := loadGroupMembersLocked(tx, groupID, lockedIDs...)
		if err != nil {
			return err
		}
		operator := members[operatorID]
		if operator == nil || operator.RoleLevel < int(constants.ManagerGroupRoleLevel) {
			return status.Error(codes.PermissionDenied, "only a current group owner or administrator may remove members")
		}
		adminIDs := make([]uint64, 0, len(memberIDs))
		for _, memberID := range memberIDs {
			if member := members[memberID]; member != nil && member.RoleLevel == int(constants.ManagerGroupRoleLevel) && member.RoleLevel < operator.RoleLevel {
				adminIDs = append(adminIDs, memberID)
			}
		}
		if err := convergeAdminReceipts(tx, groupID, adminIDs, false); err != nil {
			return err
		}
		for _, memberID := range memberIDs {
			member := members[memberID]
			if member == nil || member.RoleLevel >= operator.RoleLevel {
				continue
			}
			if err := tx.Delete(&objects.GroupMember{}, member.ID).Error; err != nil {
				return err
			}
			event := &mq.RelationChangeTransfer{
				GroupId: in.GroupId, UserId: strconv.FormatUint(memberID, 10), OperatorId: in.UserId,
			}
			version, err := emitRelationChangeInTxWithVersion(tx, l.svcCtx, constants.RelationEventGroupMemberRemoved, in.GroupId, event)
			if err != nil {
				return err
			}
			attemptRemoved = append(attemptRemoved, removedMember{userID: memberID, event: event, version: version})
		}
		removedIDs := make([]uint64, 0, len(attemptRemoved))
		for _, member := range attemptRemoved {
			removedIDs = append(removedIDs, member.userID)
		}
		if err := invalidateInvitationsByDepartingMembers(tx, groupID, removedIDs, time.Now()); err != nil {
			return err
		}
		removed = attemptRemoved
		return nil
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to remove group members")
	}
	for _, member := range removed {
		bestEffortApplyCache(l.ctx, l.svcCtx.RelationCache, member.event, int64(member.version))
		memberText := strconv.FormatUint(member.userID, 10)
		emitCommonNotify(l.ctx, l.svcCtx, &mq.CommonNotify{
			NotifyType: NotifyGroupRemoved,
			ReceiverId: memberText,
			ActorId:    in.UserId,
			BizId:      fmt.Sprintf("group.removed:%s:%s:%d", in.GroupId, memberText, time.Now().Unix()),
			GroupId:    in.GroupId,
		})
	}

	return &social.GroupKickResp{}, nil
}
