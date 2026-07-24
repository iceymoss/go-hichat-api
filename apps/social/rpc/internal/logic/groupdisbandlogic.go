package logic

import (
	"context"
	"strconv"
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

type GroupDisbandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupDisbandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupDisbandLogic {
	return &GroupDisbandLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupDisbandLogic) GroupDisband(in *social.GroupDisbandReq) (*social.GroupDisbandResp, error) {
	groupID, err := parsePositiveID(in.GroupId, "group id")
	if err != nil {
		return nil, err
	}
	actorID, err := parsePositiveID(in.UserId, "actor uid")
	if err != nil {
		return nil, err
	}
	ev := &mq.RelationChangeTransfer{
		EventType:  constants.RelationEventGroupDisbanded,
		GroupId:    in.GroupId,
		OperatorId: in.UserId,
	}
	var version uint64
	err = transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
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
		if group.CreatorUID != actorID {
			return status.Error(codes.PermissionDenied, "only the current group owner may disband the group")
		}
		if group.Status != nil && *group.Status == 1 {
			return nil
		}
		if group.Status != nil && *group.Status != 0 {
			return status.Error(codes.FailedPrecondition, "group is unavailable")
		}

		now := time.Now()
		updated := tx.Model(&objects.Group{}).Where("id = ? AND creator_uid = ? AND (status = ? OR status IS NULL)", groupID, actorID, 0).
			Updates(map[string]any{"status": 1, "updated_at": now})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			return status.Error(codes.Aborted, "group changed while being disbanded")
		}
		if err := invalidateGroupRequestsForUnavailableGroup(tx, l.svcCtx, groupID, actorID, now); err != nil {
			return err
		}
		if err := invalidateGroupInvitationsForUnavailableGroup(tx, groupID, now); err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", groupID).Delete(&objects.GroupMember{}).Error; err != nil {
			return err
		}
		ev.Timestamp = now.UnixMilli()
		var emitErr error
		version, emitErr = emitRelationChangeInTxWithVersion(tx, l.svcCtx, constants.RelationEventGroupDisbanded, in.GroupId, ev)
		return emitErr
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to disband group")
	}
	bestEffortApplyCache(l.ctx, l.svcCtx.RelationCache, ev, int64(version))

	return &social.GroupDisbandResp{}, nil
}

func invalidateGroupRequestsForUnavailableGroup(tx *gorm.DB, svcCtx *svc.ServiceContext, groupID, actorID uint64, now time.Time) error {
	var requests []objects.GroupRequest
	if err := tx.Where("group_id = ? AND handle_result = ?", groupID, groupRequestPending).Find(&requests).Error; err != nil {
		return err
	}
	actor := strconv.FormatUint(actorID, 10)
	for _, request := range requests {
		updated := tx.Model(&objects.GroupRequest{}).Where("id = ? AND handle_result = ?", request.ID, groupRequestPending).
			Updates(map[string]any{"handle_result": groupRequestInvalidated, "handle_time": now, "active_key": nil, "invalid_reason": "group disbanded"})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			continue
		}
		if err := resolveApplyReceipts(tx, receiptTypeGroup, []uint64{request.ID}, receiptInvalidated, now, ""); err != nil {
			return err
		}
		createdAt := now
		if request.ReqTime != nil {
			createdAt = *request.ReqTime
		}
		if err := createResultReceipt(tx, receiptTypeGroup, request.ID, request.ReqID, receiptInvalidated, createdAt, now, false); err != nil {
			return err
		}
		if err := emitRequestNotificationInTx(tx, svcCtx, NotifyGroupInvalidated, receiptTypeGroup, request.ID, request.ReqID, actor, groupID, receiptInvalidated, "group disbanded"); err != nil {
			return err
		}
	}
	return nil
}

func invalidateGroupInvitationsForUnavailableGroup(tx *gorm.DB, groupID uint64, now time.Time) error {
	var invitations []objects.GroupInvitation
	if err := tx.Where("group_id = ? AND status = ?", groupID, groupInvitationPending).Find(&invitations).Error; err != nil {
		return err
	}
	ids := make([]uint64, 0, len(invitations))
	for _, invitation := range invitations {
		updated := tx.Model(&objects.GroupInvitation{}).Where("id = ? AND status = ?", invitation.ID, groupInvitationPending).
			Updates(map[string]any{"status": groupInvitationInvalidated, "handled_at": now})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 1 {
			ids = append(ids, invitation.ID)
		}
	}
	return resolveInviteReceipts(tx, ids, receiptInvalidated, now, "")
}

func invalidateInvitationsByDepartingMembers(tx *gorm.DB, groupID uint64, inviterIDs []uint64, now time.Time) error {
	if len(inviterIDs) == 0 {
		return nil
	}
	var invitations []objects.GroupInvitation
	if err := tx.Where("group_id = ? AND inviter_uid IN ? AND status = ?", groupID, inviterIDs, groupInvitationPending).Find(&invitations).Error; err != nil {
		return err
	}
	ids := make([]uint64, 0, len(invitations))
	for _, invitation := range invitations {
		updated := tx.Model(&objects.GroupInvitation{}).Where("id = ? AND status = ?", invitation.ID, groupInvitationPending).
			Updates(map[string]any{"status": groupInvitationInvalidated, "handled_at": now})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 1 {
			ids = append(ids, invitation.ID)
		}
	}
	return resolveInviteReceipts(tx, ids, receiptInvalidated, now, "")
}
