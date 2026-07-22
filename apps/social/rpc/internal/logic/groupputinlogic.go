package logic

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	groupRequestPending     = 0
	groupRequestAccepted    = 1
	groupRequestRejected    = 2
	groupRequestInvalidated = 3

	groupInvitationPending     = 0
	groupInvitationAccepted    = 1
	groupInvitationRejected    = 2
	groupInvitationInvalidated = 3
	groupInvitationExpired     = 4
)

type GroupPutinLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupPutinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutinLogic {
	return &GroupPutinLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupPutinLogic) GroupPutin(in *social.GroupPutinReq) (*social.GroupPutinResp, error) {
	actor, groupID, err := validateDirectGroupRequest(in)
	if err != nil {
		return nil, err
	}
	if err := requireNormalUser(l.ctx, l.User, in.ActorUid); err != nil {
		return nil, err
	}

	var response *social.GroupPutinResp
	createdPending := false
	err = transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		response = nil
		createdPending = false
		group, err := loadNormalGroup(tx, groupID)
		if err != nil {
			return err
		}
		member, err := loadGroupMember(tx, groupID, actor)
		if err != nil {
			return err
		}
		if member != nil {
			now := time.Now()
			if err := resolvePendingGroupRequests(tx, l.ServiceContext, groupID, actor, now, 1, true, in.ActorUid); err != nil {
				return err
			}
			if err := invalidatePendingGroupInvitations(tx, groupID, actor, now); err != nil {
				return err
			}
			response = groupPutinResponse(groupID, 0, groupRequestAccepted, true, false)
			return nil
		}

		now := time.Now().UTC()
		if group.IsVerify == 0 {
			request := objects.GroupRequest{
				ReqID: strconv.FormatUint(actor, 10), GroupID: groupID, ReqMsg: in.ReqMsg, ReqTime: &now,
				JoinSource: intPtr(1), HandleResult: intPtr(groupRequestAccepted), HandleTime: &now,
				SourceType: 1, ActualJoinSource: intPtr(1),
			}
			if err := tx.Create(&request).Error; err != nil {
				return err
			}
			if err := createResultReceipt(tx, receiptTypeGroup, request.ID, in.ActorUid, receiptAccepted, now, now, true); err != nil {
				return err
			}
			memberCreated, err := createGroupMemberAndOutbox(tx, l.ServiceContext, groupID, actor, 0, 1, nil, actor)
			if err != nil {
				return err
			}
			_ = memberCreated
			if err := resolvePendingGroupRequests(tx, l.ServiceContext, groupID, actor, now, 1, true, in.ActorUid); err != nil {
				return err
			}
			if err := invalidatePendingGroupInvitations(tx, groupID, actor, now); err != nil {
				return err
			}
			response = groupPutinResponse(groupID, request.ID, groupRequestAccepted, false, false)
			return nil
		}

		activeKey := fmt.Sprintf("group:direct:%d:%d", groupID, actor)
		var existing objects.GroupRequest
		if err := tx.Where("active_key = ?", activeKey).First(&existing).Error; err == nil {
			response = groupPutinResponse(groupID, existing.ID, groupRequestPending, false, true)
			return nil
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		var latest objects.GroupRequest
		if err := tx.Where("group_id = ? AND req_id = ? AND source_type = ? AND handle_result <> ?", groupID, strconv.FormatUint(actor, 10), 1, groupRequestPending).
			Order("req_time DESC").First(&latest).Error; err == nil && latest.ReqTime != nil && now.Sub(*latest.ReqTime) < time.Minute {
			return status.Error(codes.ResourceExhausted, "group request may only be submitted once per 60 seconds")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		request := objects.GroupRequest{
			ReqID: strconv.FormatUint(actor, 10), GroupID: groupID, ReqMsg: in.ReqMsg, ReqTime: &now,
			JoinSource: intPtr(1), HandleResult: intPtr(groupRequestPending), ActiveKey: &activeKey, SourceType: 1,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&request)
		if result.Error != nil {
			return result.Error
		}
		createdPending = result.RowsAffected == 1
		if !createdPending {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("active_key = ?", activeKey).First(&request).Error; err != nil {
				return err
			}
		}
		if createdPending {
			if err := createGroupApplyReceipts(tx, l.ServiceContext, groupID, request.ID, in.ActorUid, in.ReqMsg, now); err != nil {
				return err
			}
		}
		response = groupPutinResponse(groupID, request.ID, groupRequestPending, false, !createdPending)
		return nil
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to submit group request")
	}

	return response, nil
}

// groupPutInByTokenTx keeps invite-link application writes in the caller's token transaction.
func (l *GroupPutinLogic) groupPutInByTokenTx(tx *gorm.DB, in *social.GroupPutinReq) (*social.GroupPutinResp, error) {
	actor, groupID, err := validateDirectGroupRequest(in)
	if err != nil {
		return nil, err
	}
	inviter, err := parsePositiveID(in.InviterUid, "inviter uid")
	if err != nil {
		return nil, err
	}
	group, err := loadNormalGroup(tx, groupID)
	if err != nil {
		return nil, err
	}
	member, err := loadGroupMember(tx, groupID, actor)
	if err != nil {
		return nil, err
	}
	if member != nil {
		return groupPutinResponse(groupID, 0, groupRequestAccepted, true, false), nil
	}
	inviterMember, err := loadGroupMember(tx, groupID, inviter)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	accepted := group.IsVerify == 0 || (inviterMember != nil && (inviterMember.RoleLevel == 1 || inviterMember.RoleLevel == 2))
	result := groupRequestPending
	if accepted {
		result = groupRequestAccepted
	}
	joinSource := int(in.JoinSource)
	request := objects.GroupRequest{
		ReqID: strconv.FormatUint(actor, 10), GroupID: groupID, ReqMsg: in.ReqMsg, ReqTime: &now,
		JoinSource: &joinSource, InviterUserID: &inviter, HandleResult: &result, SourceType: 2,
	}
	if accepted {
		request.HandleTime = &now
		request.ActualJoinSource = &joinSource
	}
	if err := tx.Create(&request).Error; err != nil {
		return nil, err
	}
	if accepted {
		if _, err := createGroupMemberAndOutbox(tx, l.ServiceContext, groupID, actor, 0, joinSource, &inviter, inviter); err != nil {
			return nil, err
		}
		if err := createResultReceipt(tx, receiptTypeGroup, request.ID, in.ActorUid, receiptAccepted, now, now, true); err != nil {
			return nil, err
		}
		if err := resolvePendingGroupRequests(tx, l.ServiceContext, groupID, actor, now, joinSource, true, in.ActorUid); err != nil {
			return nil, err
		}
		if err := invalidatePendingGroupInvitations(tx, groupID, actor, now); err != nil {
			return nil, err
		}
	} else if err := createGroupApplyReceipts(tx, l.ServiceContext, groupID, request.ID, in.ActorUid, in.ReqMsg, now); err != nil {
		return nil, err
	}
	return groupPutinResponse(groupID, request.ID, result, false, false), nil
}

func validateDirectGroupRequest(in *social.GroupPutinReq) (uint64, uint64, error) {
	if in.ActorUid == "" {
		return 0, 0, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if in.ReqId != "" && in.ReqId != in.ActorUid {
		return 0, 0, status.Error(codes.PermissionDenied, "actor uid does not match request uid")
	}
	actor, err := parsePositiveID(in.ActorUid, "actor uid")
	if err != nil {
		return 0, 0, err
	}
	groupID, err := parsePositiveID(in.GroupId, "group id")
	return actor, groupID, err
}

func groupPutinResponse(groupID, requestID uint64, requestStatus int, alreadyMember, alreadyPending bool) *social.GroupPutinResp {
	legacyID := int32(0)
	if groupID <= uint64(^uint32(0)>>1) {
		legacyID = int32(groupID)
	}
	return &social.GroupPutinResp{
		GroupId: legacyID, GroupIdString: strconv.FormatUint(groupID, 10), RequestId: requestID,
		IsPass: boolInt32(requestStatus == groupRequestAccepted), Status: int32(requestStatus),
		AlreadyMember: alreadyMember, AlreadyPending: alreadyPending,
	}
}

func boolInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func parsePositiveID(value, name string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, status.Errorf(codes.InvalidArgument, "%s must be a positive integer", name)
	}
	return id, nil
}

func requireNormalUser(ctx context.Context, lookup svc.UserLookup, id string) error {
	resp, err := lookup.GetUserById(ctx, &user.GetUserByIdRequest{Id: id})
	if err != nil {
		return normalizeUserLookupError(err)
	}
	if resp == nil || resp.User == nil || resp.User.Status != 1 {
		return status.Error(codes.NotFound, "user does not exist or is unavailable")
	}
	return nil
}

func loadNormalGroup(tx *gorm.DB, id uint64) (*objects.Group, error) {
	var group objects.Group
	if err := tx.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "group not found")
		}
		return nil, err
	}
	if group.Status != nil && *group.Status != 0 {
		return nil, status.Error(codes.FailedPrecondition, "group is unavailable")
	}
	return &group, nil
}

func loadGroupMember(tx *gorm.DB, groupID, userID uint64) (*objects.GroupMember, error) {
	return loadGroupMemberWithLock(tx, groupID, userID, false)
}

func loadGroupMemberLocked(tx *gorm.DB, groupID, userID uint64) (*objects.GroupMember, error) {
	return loadGroupMemberWithLock(tx, groupID, userID, true)
}

func loadGroupMembersLocked(tx *gorm.DB, groupID uint64, userIDs ...uint64) (map[uint64]*objects.GroupMember, error) {
	sorted := append([]uint64(nil), userIDs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	members := make(map[uint64]*objects.GroupMember, len(sorted))
	for _, userID := range sorted {
		if _, loaded := members[userID]; loaded {
			continue
		}
		member, err := loadGroupMemberLocked(tx, groupID, userID)
		if err != nil {
			return nil, err
		}
		members[userID] = member
	}
	return members, nil
}

func loadGroupMemberWithLock(tx *gorm.DB, groupID, userID uint64, locked bool) (*objects.GroupMember, error) {
	var member objects.GroupMember
	query := tx
	if locked && tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &member, err
}

func createGroupMemberAndOutbox(tx *gorm.DB, svcCtx *svc.ServiceContext, groupID, userID uint64, role, joinSource int, inviter *uint64, operator uint64) (bool, error) {
	now := time.Now()
	member := objects.GroupMember{
		GroupID: groupID, UserID: userID, RoleLevel: role, JoinTime: &now,
		JoinSource: &joinSource, InviterUID: inviter, OperatorUID: &operator,
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&member)
	if result.Error != nil {
		return false, result.Error
	}
	var count int64
	if err := tx.Model(&objects.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	if count != 1 {
		return false, status.Error(codes.Internal, "failed to establish group membership")
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	groupText, userText, operatorText := strconv.FormatUint(groupID, 10), strconv.FormatUint(userID, 10), strconv.FormatUint(operator, 10)
	if err := emitRelationChangeInTx(tx, svcCtx, constants.RelationEventGroupMemberAdded, groupText, &mq.RelationChangeTransfer{
		GroupId: groupText, UserId: userText, OperatorId: operatorText,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func resolvePendingGroupRequests(tx *gorm.DB, svcCtx *svc.ServiceContext, groupID, userID uint64, now time.Time, actualSource int, readResults bool, actorID string) error {
	var requests []objects.GroupRequest
	if err := tx.Where("group_id = ? AND req_id = ? AND handle_result = ?", groupID, strconv.FormatUint(userID, 10), groupRequestPending).Find(&requests).Error; err != nil {
		return err
	}
	for _, request := range requests {
		result := tx.Model(&objects.GroupRequest{}).Where("id = ? AND handle_result = ?", request.ID, groupRequestPending).
			Updates(map[string]any{"handle_result": groupRequestAccepted, "handle_time": now, "active_key": nil, "actual_join_source": actualSource})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			continue
		}
		if err := resolveApplyReceipts(tx, receiptTypeGroup, []uint64{request.ID}, receiptAccepted, now, ""); err != nil {
			return err
		}
		createdAt := now
		if request.ReqTime != nil {
			createdAt = *request.ReqTime
		}
		if err := createResultReceipt(tx, receiptTypeGroup, request.ID, request.ReqID, receiptAccepted, createdAt, now, readResults); err != nil {
			return err
		}
		if !readResults {
			if err := emitRequestNotificationInTx(tx, svcCtx, NotifyGroupAccept, receiptTypeGroup, request.ID, request.ReqID, actorID, groupID, receiptAccepted); err != nil {
				return err
			}
		}
	}
	return nil
}

func invalidatePendingGroupInvitations(tx *gorm.DB, groupID, userID uint64, now time.Time) error {
	var invitations []objects.GroupInvitation
	if err := tx.Where("group_id = ? AND invitee_uid = ? AND status = ?", groupID, userID, groupInvitationPending).Find(&invitations).Error; err != nil {
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

func transactionWithSQLiteRetry(ctx context.Context, db *gorm.DB, fn func(*gorm.DB) error) error {
	for attempt := 0; ; attempt++ {
		err := db.WithContext(ctx).Transaction(fn)
		if err == nil || db.Dialector.Name() != "sqlite" || attempt >= 9 || !isSQLiteLockError(err) {
			return err
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func isSQLiteLockError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") || strings.Contains(message, "database table is locked") || strings.Contains(message, "sqlite_busy")
}

func normalizeGroupWriteError(err error, fallback string) error {
	if status.Code(err) != codes.Unknown {
		return err
	}
	return status.Error(codes.Internal, fallback)
}
