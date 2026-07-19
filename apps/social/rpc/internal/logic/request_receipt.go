package logic

import (
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	receiptTypeFriend      = "friend"
	receiptTypeGroup       = "group"
	receiptTypeGroupInvite = "group_invite"
	receiptKindApply       = "apply"
	receiptKindResult      = "result"
	receiptKindInvite      = "invite"
	receiptPending         = 0
	receiptAccepted        = 1
	receiptRejected        = 2
	receiptInvalidated     = 3
	receiptExpired         = 4
)

type receiptCounts struct {
	Total  int32
	Apply  int32
	Result int32
	Invite int32
}

func createReceipt(tx *gorm.DB, requestType string, requestID uint64, receiverID, kind string, actionable, result int, createdAt time.Time, read bool, resolvedAt *time.Time) error {
	receipt := objects.SocialRequestReceipt{
		RequestType: requestType, RequestID: requestID, ReceiverID: receiverID, ReceiptKind: kind,
		IsActionable: actionable, Result: result, CreatedAt: createdAt, ResolvedAt: resolvedAt,
	}
	if read {
		receipt.IsRead = 1
		readAt := createdAt
		if resolvedAt != nil {
			readAt = *resolvedAt
		}
		receipt.ReadAt = &readAt
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt).Error
}

func createGroupApplyReceipts(tx *gorm.DB, svcCtx *svc.ServiceContext, groupID, requestID uint64, actorID, content string, createdAt time.Time) error {
	var members []objects.GroupMember
	if err := tx.Where("group_id = ? AND role_level IN ?", groupID, []int{1, 2}).Find(&members).Error; err != nil {
		return err
	}
	for _, member := range members {
		if err := createReceipt(tx, receiptTypeGroup, requestID, strconv.FormatUint(member.UserID, 10), receiptKindApply, 1, receiptPending, createdAt, false, nil); err != nil {
			return err
		}
		if err := emitRequestNotificationInTx(tx, svcCtx, NotifyGroupApply, receiptTypeGroup, requestID, strconv.FormatUint(member.UserID, 10), actorID, groupID, receiptPending, content); err != nil {
			return err
		}
	}
	return nil
}

func resolveApplyReceipts(tx *gorm.DB, requestType string, requestIDs []uint64, result int, now time.Time, readActor string) error {
	if len(requestIDs) == 0 {
		return nil
	}
	updates := map[string]any{"is_actionable": 0, "result": result, "resolved_at": now}
	if err := tx.Model(&objects.SocialRequestReceipt{}).
		Where("request_type = ? AND request_id IN ? AND receipt_kind = ? AND result <> ?", requestType, requestIDs, receiptKindApply, receiptInvalidated).
		Updates(updates).Error; err != nil {
		return err
	}
	if readActor != "" {
		return tx.Model(&objects.SocialRequestReceipt{}).
			Where("request_type = ? AND request_id IN ? AND receipt_kind = ? AND receiver_id = ?", requestType, requestIDs, receiptKindApply, readActor).
			Updates(map[string]any{"is_read": 1, "read_at": now}).Error
	}
	return nil
}

func createResultReceipt(tx *gorm.DB, requestType string, requestID uint64, receiverID string, result int, createdAt, now time.Time, read bool) error {
	return createReceipt(tx, requestType, requestID, receiverID, receiptKindResult, 0, result, createdAt, read, &now)
}

func resolveInviteReceipts(tx *gorm.DB, invitationIDs []uint64, result int, now time.Time, readActor string) error {
	if len(invitationIDs) == 0 {
		return nil
	}
	if err := tx.Model(&objects.SocialRequestReceipt{}).
		Where("request_type = ? AND request_id IN ? AND receipt_kind = ?", receiptTypeGroupInvite, invitationIDs, receiptKindInvite).
		Updates(map[string]any{"is_actionable": 0, "result": result, "resolved_at": now}).Error; err != nil {
		return err
	}
	if readActor != "" {
		return tx.Model(&objects.SocialRequestReceipt{}).
			Where("request_type = ? AND request_id IN ? AND receipt_kind = ? AND receiver_id = ?", receiptTypeGroupInvite, invitationIDs, receiptKindInvite, readActor).
			Updates(map[string]any{"is_read": 1, "read_at": now}).Error
	}
	return nil
}

func markReceiptsRead(tx *gorm.DB, receiverID, requestType string, requestIDs []uint64) error {
	query := tx.Model(&objects.SocialRequestReceipt{}).
		Where("receiver_id = ? AND request_type = ? AND is_read = ?", receiverID, requestType, 0)
	if len(requestIDs) > 0 {
		query = query.Where("request_id IN ?", requestIDs)
	}
	now := time.Now()
	return query.Updates(map[string]any{"is_read": 1, "read_at": now}).Error
}

func countUnreadReceipts(tx *gorm.DB, receiverID string, group bool) (receiptCounts, error) {
	type row struct {
		RequestType string
		ReceiptKind string
		Count       int32
	}
	query := tx.Model(&objects.SocialRequestReceipt{}).
		Select("request_type, receipt_kind, COUNT(*) AS count").
		Where("receiver_id = ? AND is_read = ?", receiverID, 0)
	if group {
		query = query.Where("request_type IN ?", []string{receiptTypeGroup, receiptTypeGroupInvite})
	} else {
		query = query.Where("request_type = ?", receiptTypeFriend)
	}
	var rows []row
	if err := query.Group("request_type, receipt_kind").Scan(&rows).Error; err != nil {
		return receiptCounts{}, err
	}
	var counts receiptCounts
	for _, item := range rows {
		switch item.ReceiptKind {
		case receiptKindApply:
			counts.Apply += item.Count
		case receiptKindResult:
			counts.Result += item.Count
		case receiptKindInvite:
			counts.Invite += item.Count
		}
	}
	counts.Total = counts.Apply + counts.Result + counts.Invite
	return counts, nil
}
