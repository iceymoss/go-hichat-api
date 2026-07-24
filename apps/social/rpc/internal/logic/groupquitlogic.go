package logic

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	groupID, parseErr := parsePositiveID(in.GroupId, "group id")
	if parseErr != nil {
		return nil, parseErr
	}
	userID, parseErr := parsePositiveID(in.UserId, "user id")
	if parseErr != nil {
		return nil, parseErr
	}
	event := &mq.RelationChangeTransfer{GroupId: in.GroupId, UserId: in.UserId, OperatorId: in.UserId}
	var version uint64
	err := transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
		version = 0
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
		member, err := loadGroupMemberLocked(tx, groupID, userID)
		if err != nil {
			return err
		}
		if member == nil {
			return status.Error(codes.NotFound, "group member not found")
		}
		if member.RoleLevel == int(constants.CreatorGroupRoleLevel) {
			return status.Error(codes.FailedPrecondition, "owner must transfer ownership or disband the group before leaving")
		}
		if member.RoleLevel == int(constants.ManagerGroupRoleLevel) {
			if err := convergeAdminReceipts(tx, groupID, []uint64{userID}, false); err != nil {
				return err
			}
		}
		if err := tx.Delete(&objects.GroupMember{}, member.ID).Error; err != nil {
			return err
		}
		if err := invalidateInvitationsByDepartingMembers(tx, groupID, []uint64{userID}, time.Now()); err != nil {
			return err
		}
		var emitErr error
		version, emitErr = emitRelationChangeInTxWithVersion(tx, l.svcCtx, constants.RelationEventGroupMemberRemoved, in.GroupId, event)
		return emitErr
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to leave group")
	}
	bestEffortApplyCache(l.ctx, l.svcCtx.RelationCache, event, int64(version))

	return &social.GroupQuitResp{}, nil
}
