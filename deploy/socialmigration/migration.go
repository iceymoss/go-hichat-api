package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/zeromicro/go-zero/core/jsonx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Report struct {
	Version                  string                  `json:"version"`
	AlreadyApplied           bool                    `json:"alreadyApplied"`
	DataAlreadyApplied       bool                    `json:"dataAlreadyApplied"`
	ReadTimeApproximation    string                  `json:"readTimeApproximation"`
	FriendRequestDuplicates  []RequestCleanupFinding `json:"friendRequestDuplicates"`
	GroupRequestDuplicates   []RequestCleanupFinding `json:"groupRequestDuplicates"`
	FriendMerges             []FriendMergeFinding    `json:"friendMerges"`
	RecoverableInvitations   []InvitationFinding     `json:"recoverableInvitations"`
	NonUserInvitations       []InvitationFinding     `json:"nonUserInvitations"`
	AmbiguousInvitations     []InvitationFinding     `json:"ambiguousInvitations"`
	OneWayFriends            []OneWayFriendFinding   `json:"oneWayFriends"`
	OneWayPendingRequests    []OneWayPendingFinding  `json:"oneWayPendingRequests"`
	ResolvedFriendRequestIDs []uint64                `json:"resolvedFriendRequestIds"`
	ResolvedGroupRequestIDs  []uint64                `json:"resolvedGroupRequestIds"`
}

type RequestCleanupFinding struct {
	UserID     uint64   `json:"userId,omitempty"`
	RequestUID uint64   `json:"requestUserId,omitempty"`
	GroupID    uint64   `json:"groupId,omitempty"`
	Applicant  string   `json:"applicant,omitempty"`
	KeepID     uint64   `json:"keepId"`
	ChangedIDs []uint64 `json:"changedIds"`
	Rule       string   `json:"rule"`
}

type InvitationFinding struct {
	RequestID uint64 `json:"requestId"`
	GroupID   uint64 `json:"groupId"`
	InviterID uint64 `json:"inviterId,omitempty"`
	InviteeID uint64 `json:"inviteeId,omitempty"`
	Status    int    `json:"status"`
	Reason    string `json:"reason"`
}

type FriendMergeFinding struct {
	UserID            uint64   `json:"userId"`
	FriendUID         uint64   `json:"friendUid"`
	AffectedUserIDs   []uint64 `json:"affectedUserIds"`
	KeepID            uint64   `json:"keepId"`
	DeleteIDs         []uint64 `json:"deleteIds"`
	Rule              string   `json:"rule"`
	Conflicts         []string `json:"conflicts"`
	Remark            string   `json:"remark"`
	AddSource         *int     `json:"addSource,omitempty"`
	Blacklisted       bool     `json:"blacklisted"`
	MomentsPermission int      `json:"momentsPermission"`
	NotifyEnabled     bool     `json:"notifyEnabled"`
	Pinned            bool     `json:"pinned"`
	Muted             bool     `json:"muted"`
	FriendTags        []string `json:"friendTags"`
}

type OneWayFriendFinding struct {
	UserID    uint64 `json:"userId"`
	FriendUID uint64 `json:"friendUid"`
	RecordID  uint64 `json:"recordId"`
}

type OneWayPendingFinding struct {
	RequestID uint64 `json:"requestId"`
	UserID    uint64 `json:"userId"`
	FriendUID uint64 `json:"friendUid"`
	Reason    string `json:"reason"`
}

type migrationRecord struct {
	ID          uint64    `gorm:"primaryKey"`
	Version     string    `gorm:"size:64;not null;uniqueIndex"`
	Description string    `gorm:"size:255;not null"`
	AppliedAt   time.Time `gorm:"not null"`
}

func (migrationRecord) TableName() string { return "schema_migrations" }

type friendRequestColumns struct {
	ActiveKey *string `gorm:"column:active_key;size:160"`
	Status    *int    `gorm:"column:status"`
	Remark    string  `gorm:"column:remark;size:64;not null;default:''"`
}

func (friendRequestColumns) TableName() string { return "friend_requests" }

type groupRequestColumns struct {
	ActiveKey          *string `gorm:"column:active_key;size:160"`
	SourceType         int     `gorm:"column:source_type;not null;default:1"`
	SourceInvitationID *uint64 `gorm:"column:source_invitation_id"`
	ActualJoinSource   *int    `gorm:"column:actual_join_source"`
	InvalidReason      string  `gorm:"column:invalid_reason;size:128;not null;default:''"`
	HandleMsg          string  `gorm:"column:handle_msg;size:255;not null;default:''"`
}

func (groupRequestColumns) TableName() string { return "group_requests" }

func newReport() Report {
	return Report{
		Version: migrationVersion, ReadTimeApproximation: "legacy read flags have no timestamp; read_at is migration time",
		FriendRequestDuplicates: []RequestCleanupFinding{}, GroupRequestDuplicates: []RequestCleanupFinding{},
		FriendMerges: []FriendMergeFinding{}, RecoverableInvitations: []InvitationFinding{},
		NonUserInvitations: []InvitationFinding{}, AmbiguousInvitations: []InvitationFinding{},
		OneWayFriends: []OneWayFriendFinding{}, OneWayPendingRequests: []OneWayPendingFinding{},
		ResolvedFriendRequestIDs: []uint64{}, ResolvedGroupRequestIDs: []uint64{},
	}
}

func Migrate(ctx context.Context, db *gorm.DB, driver string, now time.Time) (Report, error) {
	report := newReport()
	if driver != "sqlite" && driver != "mysql" && driver != "postgres" {
		return report, fmt.Errorf("unsupported driver %q", driver)
	}
	db = db.WithContext(ctx)
	if err := db.AutoMigrate(&migrationRecord{}); err != nil {
		return report, fmt.Errorf("create schema migrations table: %w", err)
	}
	var count int64
	if err := db.Model(&migrationRecord{}).Where("version = ?", migrationVersion).Count(&count).Error; err != nil {
		return report, fmt.Errorf("check migration version: %w", err)
	}
	if count > 0 {
		report.AlreadyApplied = true
		if err := repairInvitationReceiptResults(db, now); err != nil {
			return report, err
		}
		if err := ensureGroupHandleMsg(db, now); err != nil {
			return report, err
		}
		return report, validateSchema(db)
	}

	// DDL is a repeatable phase because MySQL commits it implicitly.
	if err := ensureSchema(db); err != nil {
		return report, err
	}
	var dataCount int64
	if err := db.Model(&migrationRecord{}).Where("version = ?", migrationDataVersion).Count(&dataCount).Error; err != nil {
		return report, fmt.Errorf("check migration data phase: %w", err)
	}
	if dataCount == 0 {
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := migrateData(tx, driver, now, &report); err != nil {
				return err
			}
			phase := migrationRecord{Version: migrationDataVersion, Description: "social request reliability data cleanup", AppliedAt: now}
			if err := tx.Create(&phase).Error; err != nil {
				return fmt.Errorf("record migration data phase: %w", err)
			}
			return nil
		}); err != nil {
			return report, err
		}
	} else {
		report.DataAlreadyApplied = true
	}
	// Unique indexes are deliberately created only after cleanup commits. A
	// partial index phase is safe to rerun and does not repeat destructive DML.
	if err := createAndValidateIndexes(db); err != nil {
		return report, err
	}
	if err := repairInvitationReceiptResults(db, now); err != nil {
		return report, err
	}
	if err := ensureGroupHandleMsg(db, now); err != nil {
		return report, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		record := migrationRecord{Version: migrationVersion, Description: "social request interaction reliability schema", AppliedAt: now}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; err != nil {
			return fmt.Errorf("record migration version: %w", err)
		}
		return nil
	}); err != nil {
		return report, err
	}
	return report, validateSchema(db)
}

func ensureGroupHandleMsg(db *gorm.DB, now time.Time) error {
	var applied int64
	if err := db.Model(&migrationRecord{}).Where("version = ?", groupHandleMsgVersion).Count(&applied).Error; err != nil {
		return err
	}
	if applied > 0 {
		return nil
	}
	if !db.Migrator().HasColumn(&groupRequestColumns{}, "HandleMsg") {
		if err := db.Migrator().AddColumn(&groupRequestColumns{}, "HandleMsg"); err != nil {
			return fmt.Errorf("add group_requests.handle_msg: %w", err)
		}
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&migrationRecord{Version: groupHandleMsgVersion, Description: "add group request handle message", AppliedAt: now}).Error
}

func repairInvitationReceiptResults(db *gorm.DB, now time.Time) error {
	var applied int64
	if err := db.Model(&migrationRecord{}).Where("version = ?", receiptResultFixVersion).Count(&applied).Error; err != nil {
		return fmt.Errorf("check receipt result fix: %w", err)
	}
	if applied > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for invitationStatus, receiptResult := range map[int]int{3: 4, 4: 3} {
			var ids []uint64
			if err := tx.Model(&objects.GroupInvitation{}).Where("status = ?", invitationStatus).Pluck("id", &ids).Error; err != nil {
				return fmt.Errorf("list invitation receipt fixes: %w", err)
			}
			if len(ids) == 0 {
				continue
			}
			if err := tx.Model(&objects.SocialRequestReceipt{}).
				Where("request_type = ? AND receipt_kind = ? AND request_id IN ?", "group_invite", "invite", ids).
				Update("result", receiptResult).Error; err != nil {
				return fmt.Errorf("repair invitation receipt results: %w", err)
			}
		}
		record := migrationRecord{Version: receiptResultFixVersion, Description: "repair group invitation receipt result enum", AppliedAt: now}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; err != nil {
			return fmt.Errorf("record receipt result fix: %w", err)
		}
		return nil
	})
}

func ensureSchema(db *gorm.DB) error {
	for _, field := range []string{"Status", "Remark", "ActiveKey"} {
		if !db.Migrator().HasColumn(&friendRequestColumns{}, field) {
			if err := db.Migrator().AddColumn(&friendRequestColumns{}, field); err != nil {
				return fmt.Errorf("add friend_requests.%s: %w", field, err)
			}
		}
	}
	for _, field := range []string{"ActiveKey", "SourceType", "SourceInvitationID", "ActualJoinSource", "InvalidReason", "HandleMsg"} {
		if !db.Migrator().HasColumn(&groupRequestColumns{}, field) {
			if err := db.Migrator().AddColumn(&groupRequestColumns{}, field); err != nil {
				return fmt.Errorf("add group_requests.%s: %w", field, err)
			}
		}
	}
	if err := db.AutoMigrate(&objects.GroupInvitation{}, &objects.SocialRequestReceipt{}, &objects.SocialNotificationOutbox{}); err != nil {
		return fmt.Errorf("create social reliability tables: %w", err)
	}
	return nil
}

func migrateData(db *gorm.DB, driver string, now time.Time, report *Report) error {
	if err := classifyInvitations(db, report); err != nil {
		return err
	}
	if len(report.AmbiguousInvitations) > 0 {
		return fmt.Errorf("%d ambiguous legacy invitations require manual review", len(report.AmbiguousInvitations))
	}
	if err := migrateRecoverableInvitations(db, now, report.RecoverableInvitations); err != nil {
		return err
	}
	if err := cleanPendingRequests(db, report); err != nil {
		return err
	}
	if err := mergeDuplicateFriends(db, report); err != nil {
		return err
	}
	if err := resolveExistingRelations(db, driver, now, report); err != nil {
		return err
	}
	if err := reportOneWayFriends(db, report); err != nil {
		return err
	}
	if err := backfillActiveKeys(db, driver); err != nil {
		return err
	}
	return backfillReceipts(db, now)
}

func classifyInvitations(db *gorm.DB, report *Report) error {
	type row struct {
		ID, GroupID   uint64
		ReqID         string
		JoinSource    *int
		InviterUserID *uint64
		HandleResult  *int
	}
	var rows []row
	if err := db.Table("group_requests").Select("id, group_id, req_id, join_source, inviter_user_id, handle_result").Where("join_source IN ? OR inviter_user_id IS NOT NULL", []int{2, 3}).Order("id").Scan(&rows).Error; err != nil {
		return fmt.Errorf("classify legacy invitations: %w", err)
	}
	checkUsers := db.Migrator().HasTable(&objects.User{})
	checkGroups := db.Migrator().HasTable(&objects.Group{})
	for _, row := range rows {
		finding := InvitationFinding{RequestID: row.ID, GroupID: row.GroupID}
		invitee, parseErr := strconv.ParseUint(row.ReqID, 10, 64)
		if parseErr == nil {
			finding.InviteeID = invitee
		}
		if row.InviterUserID != nil {
			finding.InviterID = *row.InviterUserID
		}
		status, statusOK := legacyStatus(row.HandleResult)
		finding.Status = status
		reason := ""
		switch {
		case row.JoinSource != nil && *row.JoinSource == 3 && statusOK:
			finding.Reason = "join_source=3 is an invite link/QR flow"
			report.NonUserInvitations = append(report.NonUserInvitations, finding)
			continue
		case row.JoinSource == nil || *row.JoinSource != 2:
			reason = "member-invite source is missing or unknown"
		case row.InviterUserID == nil || *row.InviterUserID == 0:
			reason = "inviter is missing"
		case parseErr != nil || invitee == 0:
			reason = "req_id is not a numeric uint invitee ID"
		case !statusOK:
			reason = "handle_result is outside the explicit 0=pending,1=accepted,2=rejected mapping"
		}
		if reason == "" && checkGroups {
			exists, err := recordExists(db, "groups", row.GroupID)
			if err != nil {
				return fmt.Errorf("check invitation group %d: %w", row.GroupID, err)
			}
			if !exists {
				reason = "group does not exist"
			}
		}
		if reason == "" && checkUsers {
			inviterExists, err := recordExists(db, "users", *row.InviterUserID)
			if err != nil {
				return fmt.Errorf("check inviter %d: %w", *row.InviterUserID, err)
			}
			inviteeExists, err := recordExists(db, "users", invitee)
			if err != nil {
				return fmt.Errorf("check invitee %d: %w", invitee, err)
			}
			if !inviterExists {
				reason = "inviter does not exist"
			} else if !inviteeExists {
				reason = "invitee does not exist"
			}
		}
		if reason != "" {
			finding.Reason = reason
			report.AmbiguousInvitations = append(report.AmbiguousInvitations, finding)
			continue
		}
		finding.Reason = "recoverable member invitation; legacy schema cannot prove invitee confirmation"
		report.RecoverableInvitations = append(report.RecoverableInvitations, finding)
	}
	return nil
}

func legacyStatus(value *int) (int, bool) {
	if value == nil || *value == 0 {
		return 0, true
	}
	if *value == 1 || *value == 2 {
		return *value, true
	}
	return *value, false
}

func recordExists(db *gorm.DB, table string, id uint64) (bool, error) {
	var count int64
	if err := db.Table(table).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func migrateRecoverableInvitations(db *gorm.DB, now time.Time, findings []InvitationFinding) error {
	type legacy struct {
		ID, GroupID   uint64
		ReqID         string
		ReqMsg        string
		ReqTime       *time.Time
		InviterUserID uint64
		HandleResult  *int
		HandleTime    *time.Time
	}
	for _, finding := range findings {
		var row legacy
		if err := db.Table("group_requests").Where("id = ?", finding.RequestID).First(&row).Error; err != nil {
			return fmt.Errorf("read recoverable invitation %d: %w", finding.RequestID, err)
		}
		createdAt := now
		if row.ReqTime != nil {
			createdAt = *row.ReqTime
		}
		status, _ := legacyStatus(row.HandleResult)
		expiresAt := createdAt.Add(7 * 24 * time.Hour)
		member, err := groupMemberExists(db, row.GroupID, finding.InviteeID)
		if err != nil {
			return fmt.Errorf("check invitation %d membership: %w", row.ID, err)
		}
		if member {
			status = 1
		}
		var handledAt *time.Time
		if status != 0 {
			handledAt = row.HandleTime
			if handledAt == nil {
				handledAt = &now
			}
		} else if !expiresAt.After(now) {
			status = 4
			handledAt = &now
		}
		invitation := objects.GroupInvitation{ID: row.ID, GroupID: row.GroupID, InviterUID: row.InviterUserID, InviteeUID: finding.InviteeID, Message: row.ReqMsg, Status: status, CreatedAt: createdAt, HandledAt: handledAt, ExpiresAt: expiresAt}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&invitation).Error; err != nil {
			return fmt.Errorf("migrate invitation %d: %w", row.ID, err)
		}
		var stored objects.GroupInvitation
		if err := db.First(&stored, row.ID).Error; err != nil {
			return fmt.Errorf("verify migrated invitation %d: %w", row.ID, err)
		}
		if stored.GroupID != invitation.GroupID || stored.InviterUID != invitation.InviterUID || stored.InviteeUID != invitation.InviteeUID {
			return fmt.Errorf("group invitation id %d already exists with conflicting participants", row.ID)
		}
		updates := map[string]any{"source_type": 2, "source_invitation_id": nil, "active_key": nil}
		if member {
			updates["handle_result"] = 1
			updates["handle_time"] = handledAt
			updates["actual_join_source"] = 2
		} else if finding.Status == 0 {
			updates["handle_result"] = 3
			updates["handle_time"] = now
			updates["invalid_reason"] = "migration_unconfirmed_invitation"
		}
		if err := db.Table("group_requests").Where("id = ?", row.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("close legacy invitation request %d: %w", row.ID, err)
		}
	}
	return nil
}

func groupMemberExists(db *gorm.DB, groupID, userID uint64) (bool, error) {
	var count int64
	if err := db.Table("group_members").Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func cleanPendingRequests(db *gorm.DB, report *Report) error {
	type friendPair struct{ UserID, ReqUID uint64 }
	var friendPairs []friendPair
	if err := db.Table("friend_requests").Select("user_id, req_uid").Where("COALESCE(handle_result, 0) = 0 AND COALESCE(status, 1) = 1").Group("user_id, req_uid").Having("COUNT(*) > 1").Scan(&friendPairs).Error; err != nil {
		return fmt.Errorf("find duplicate friend pending: %w", err)
	}
	for _, pair := range friendPairs {
		var ids []uint64
		if err := db.Table("friend_requests").Where("user_id = ? AND req_uid = ? AND COALESCE(handle_result, 0) = 0 AND COALESCE(status, 1) = 1", pair.UserID, pair.ReqUID).Order("req_time DESC, id DESC").Pluck("id", &ids).Error; err != nil {
			return fmt.Errorf("list duplicate friend pending: %w", err)
		}
		finding := RequestCleanupFinding{UserID: pair.UserID, RequestUID: pair.ReqUID, KeepID: ids[0], ChangedIDs: append([]uint64(nil), ids[1:]...), Rule: "keep latest req_time then largest id; older rows become ignored hidden history"}
		if err := db.Table("friend_requests").Where("id IN ?", finding.ChangedIDs).Updates(map[string]any{"handle_result": 3, "status": 2, "active_key": nil}).Error; err != nil {
			return fmt.Errorf("ignore duplicate friend pending: %w", err)
		}
		report.FriendRequestDuplicates = append(report.FriendRequestDuplicates, finding)
	}
	type groupPair struct {
		GroupID uint64
		ReqID   string
	}
	var groupPairs []groupPair
	if err := db.Table("group_requests").Select("group_id, req_id").Where("COALESCE(handle_result, 0) = 0 AND COALESCE(source_type, 1) = 1").Group("group_id, req_id").Having("COUNT(*) > 1").Scan(&groupPairs).Error; err != nil {
		return fmt.Errorf("find duplicate group pending: %w", err)
	}
	for _, pair := range groupPairs {
		var ids []uint64
		if err := db.Table("group_requests").Where("group_id = ? AND req_id = ? AND COALESCE(handle_result, 0) = 0 AND COALESCE(source_type, 1) = 1", pair.GroupID, pair.ReqID).Order("CASE WHEN req_time IS NULL THEN 1 ELSE 0 END, req_time DESC, id DESC").Pluck("id", &ids).Error; err != nil {
			return fmt.Errorf("list duplicate group pending: %w", err)
		}
		finding := RequestCleanupFinding{GroupID: pair.GroupID, Applicant: pair.ReqID, KeepID: ids[0], ChangedIDs: append([]uint64(nil), ids[1:]...), Rule: "keep latest non-null req_time then largest id; older rows become internal invalidated history"}
		if err := db.Table("group_requests").Where("id IN ?", finding.ChangedIDs).Updates(map[string]any{"handle_result": 3, "invalid_reason": "migration_duplicate_pending", "active_key": nil}).Error; err != nil {
			return fmt.Errorf("invalidate duplicate group pending: %w", err)
		}
		report.GroupRequestDuplicates = append(report.GroupRequestDuplicates, finding)
	}
	return nil
}

func mergeDuplicateFriends(db *gorm.DB, report *Report) error {
	type pair struct{ UserID, FriendUID uint64 }
	var pairs []pair
	if err := db.Table("friends").Select("user_id, friend_uid").Group("user_id, friend_uid").Having("COUNT(*) > 1").Scan(&pairs).Error; err != nil {
		return fmt.Errorf("find duplicate friends: %w", err)
	}
	for _, pair := range pairs {
		type row struct {
			ID                uint64
			Remark            string
			AddSource         *int
			Blacklisted       bool
			MomentsPermission int
			NotifyEnabled     bool
			Pinned            bool
			Muted             bool
			FriendTags        string
			CreatedAt         *time.Time
		}
		var rows []row
		if err := db.Table("friends").Where("user_id = ? AND friend_uid = ?", pair.UserID, pair.FriendUID).Order("CASE WHEN created_at IS NULL THEN 1 ELSE 0 END, created_at ASC, id ASC").Find(&rows).Error; err != nil {
			return fmt.Errorf("read duplicate friend metadata: %w", err)
		}
		finding := FriendMergeFinding{UserID: pair.UserID, FriendUID: pair.FriendUID, AffectedUserIDs: []uint64{pair.UserID, pair.FriendUID}, KeepID: rows[0].ID, Rule: "keep earliest relation; retain earliest non-empty scalar; merge stricter settings and JSON tag union", Conflicts: []string{}, FriendTags: []string{}, NotifyEnabled: true}
		tagSet := make(map[string]struct{})
		for index, row := range rows {
			if index > 0 {
				finding.DeleteIDs = append(finding.DeleteIDs, row.ID)
			}
			if row.Remark != "" {
				if finding.Remark == "" {
					finding.Remark = row.Remark
				} else if finding.Remark != row.Remark {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("remark row %d=%q kept %q", row.ID, row.Remark, finding.Remark))
				}
			}
			if row.AddSource != nil {
				if finding.AddSource == nil {
					value := *row.AddSource
					finding.AddSource = &value
				} else if *finding.AddSource != *row.AddSource {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("add_source row %d=%d kept %d", row.ID, *row.AddSource, *finding.AddSource))
				}
			}
			if index > 0 {
				first := rows[0]
				if first.Blacklisted != row.Blacklisted {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("blacklisted row %d=%t merged with OR", row.ID, row.Blacklisted))
				}
				if first.MomentsPermission != row.MomentsPermission {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("moments_permission row %d=%d merged with max", row.ID, row.MomentsPermission))
				}
				if first.NotifyEnabled != row.NotifyEnabled {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("notify_enabled row %d=%t merged with AND", row.ID, row.NotifyEnabled))
				}
				if first.Pinned != row.Pinned {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("pinned row %d=%t merged with OR", row.ID, row.Pinned))
				}
				if first.Muted != row.Muted {
					finding.Conflicts = append(finding.Conflicts, fmt.Sprintf("muted row %d=%t merged with OR", row.ID, row.Muted))
				}
			}
			finding.Blacklisted = finding.Blacklisted || row.Blacklisted
			if row.MomentsPermission > finding.MomentsPermission {
				finding.MomentsPermission = row.MomentsPermission
			}
			finding.NotifyEnabled = finding.NotifyEnabled && row.NotifyEnabled
			finding.Pinned = finding.Pinned || row.Pinned
			finding.Muted = finding.Muted || row.Muted
			if row.FriendTags != "" {
				var tags []string
				if err := jsonx.Unmarshal([]byte(row.FriendTags), &tags); err != nil {
					return fmt.Errorf("friend relation %d has invalid friend_tags JSON: %w", row.ID, err)
				}
				for _, tag := range tags {
					if _, exists := tagSet[tag]; !exists {
						tagSet[tag] = struct{}{}
						finding.FriendTags = append(finding.FriendTags, tag)
					}
				}
			}
		}
		tags, err := jsonx.Marshal(finding.FriendTags)
		if err != nil {
			return fmt.Errorf("encode merged friend tags: %w", err)
		}
		updates := map[string]any{"remark": finding.Remark, "add_source": finding.AddSource, "blacklisted": finding.Blacklisted, "moments_permission": finding.MomentsPermission, "notify_enabled": finding.NotifyEnabled, "pinned": finding.Pinned, "muted": finding.Muted, "friend_tags": string(tags)}
		if err := db.Table("friends").Where("id = ?", finding.KeepID).Updates(updates).Error; err != nil {
			return fmt.Errorf("merge duplicate friend metadata: %w", err)
		}
		if err := db.Table("friends").Where("id IN ?", finding.DeleteIDs).Delete(nil).Error; err != nil {
			return fmt.Errorf("delete duplicate friends: %w", err)
		}
		report.FriendMerges = append(report.FriendMerges, finding)
	}
	return nil
}

func resolveExistingRelations(db *gorm.DB, driver string, now time.Time, report *Report) error {
	type friendRequest struct{ ID, UserID, ReqUID uint64 }
	var requests []friendRequest
	if err := db.Table("friend_requests").Select("id, user_id, req_uid").Where("COALESCE(handle_result, 0) = 0").Find(&requests).Error; err != nil {
		return fmt.Errorf("read pending friend requests: %w", err)
	}
	for _, request := range requests {
		forward, err := friendRelationExists(db, request.UserID, request.ReqUID)
		if err != nil {
			return fmt.Errorf("check forward friendship for request %d: %w", request.ID, err)
		}
		reverse, err := friendRelationExists(db, request.ReqUID, request.UserID)
		if err != nil {
			return fmt.Errorf("check reverse friendship for request %d: %w", request.ID, err)
		}
		if forward && reverse {
			if err := db.Table("friend_requests").Where("id = ?", request.ID).Updates(map[string]any{"handle_result": 1, "active_key": nil, "handled_at": now}).Error; err != nil {
				return fmt.Errorf("resolve existing bidirectional friendship: %w", err)
			}
			report.ResolvedFriendRequestIDs = append(report.ResolvedFriendRequestIDs, request.ID)
		} else if forward || reverse {
			report.OneWayPendingRequests = append(report.OneWayPendingRequests, OneWayPendingFinding{RequestID: request.ID, UserID: request.UserID, FriendUID: request.ReqUID, Reason: "only one directed friendship exists; request remains pending"})
		}
	}
	memberUser := "CAST(gm.user_id AS VARCHAR(64))"
	if driver == "mysql" {
		memberUser = "CAST(gm.user_id AS CHAR)"
	}
	type idRow struct{ ID uint64 }
	var memberRequests []idRow
	query := `SELECT gr.id FROM group_requests gr WHERE COALESCE(gr.handle_result, 0) = 0 AND EXISTS (SELECT 1 FROM group_members gm WHERE gm.group_id = gr.group_id AND ` + memberUser + ` = gr.req_id)`
	if err := db.Raw(query).Scan(&memberRequests).Error; err != nil {
		return fmt.Errorf("find pending requests for existing members: %w", err)
	}
	for _, request := range memberRequests {
		if err := db.Table("group_requests").Where("id = ?", request.ID).Updates(map[string]any{"handle_result": 1, "active_key": nil, "handle_time": now, "actual_join_source": 1}).Error; err != nil {
			return fmt.Errorf("resolve pending request for existing member: %w", err)
		}
		report.ResolvedGroupRequestIDs = append(report.ResolvedGroupRequestIDs, request.ID)
	}
	return nil
}

func friendRelationExists(db *gorm.DB, userID, friendUID uint64) (bool, error) {
	var count int64
	if err := db.Table("friends").Where("user_id = ? AND friend_uid = ?", userID, friendUID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func reportOneWayFriends(db *gorm.DB, report *Report) error {
	return db.Table("friends AS f").Select("f.id AS record_id, f.user_id, f.friend_uid").Where("NOT EXISTS (?)", db.Table("friends AS reverse").Select("1").Where("reverse.user_id = f.friend_uid AND reverse.friend_uid = f.user_id")).Order("f.user_id, f.friend_uid, f.id").Scan(&report.OneWayFriends).Error
}

func backfillActiveKeys(db *gorm.DB, driver string) error {
	friendExpression := "'friend:' || user_id || ':' || req_uid"
	groupExpression := "'group:direct:' || group_id || ':' || req_id"
	if driver == "mysql" {
		friendExpression = "CONCAT('friend:', user_id, ':', req_uid)"
		groupExpression = "CONCAT('group:direct:', group_id, ':', req_id)"
	}
	if err := db.Exec(`UPDATE friend_requests SET active_key = ` + friendExpression + ` WHERE COALESCE(handle_result, 0) = 0 AND COALESCE(status, 1) = 1 AND active_key IS NULL`).Error; err != nil {
		return fmt.Errorf("backfill friend active keys: %w", err)
	}
	if err := db.Exec(`UPDATE group_requests SET active_key = ` + groupExpression + ` WHERE COALESCE(handle_result, 0) = 0 AND COALESCE(source_type, 1) = 1 AND active_key IS NULL`).Error; err != nil {
		return fmt.Errorf("backfill group active keys: %w", err)
	}
	return nil
}

func backfillReceipts(db *gorm.DB, now time.Time) error {
	var invitations []objects.GroupInvitation
	if err := db.Find(&invitations).Error; err != nil {
		return fmt.Errorf("read group invitations for receipts: %w", err)
	}
	for _, invitation := range invitations {
		actionable := boolInt(invitation.Status == 0)
		result := invitation.Status
		if result == 3 {
			result = 4
		} else if result == 4 {
			result = 3
		}
		inviteReceipt := receipt("group_invite", invitation.ID, strconv.FormatUint(invitation.InviteeUID, 10), "invite", 0, actionable, result, invitation.CreatedAt, invitation.HandledAt, now)
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&inviteReceipt).Error; err != nil {
			return fmt.Errorf("backfill group invitation receipt: %w", err)
		}
	}
	type friendRow struct {
		ID, UserID, ReqUID       uint64
		HandleResult             *int
		ReceiverRead, SenderRead int
		ReqTime                  time.Time
		HandledAt                *time.Time
		Status                   *int
	}
	var friends []friendRow
	if err := db.Table("friend_requests").Find(&friends).Error; err != nil {
		return fmt.Errorf("read friend requests for receipts: %w", err)
	}
	for _, row := range friends {
		result, _ := legacyStatus(row.HandleResult)
		hidden := result == 3 || (row.Status != nil && *row.Status == 2)
		isRead := normalizeRead(row.ReceiverRead)
		if hidden {
			isRead = 1
		}
		apply := receipt("friend", row.ID, strconv.FormatUint(row.ReqUID, 10), "apply", isRead, boolInt(result == 0), result, row.ReqTime, row.HandledAt, now)
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&apply).Error; err != nil {
			return fmt.Errorf("backfill friend apply receipt: %w", err)
		}
		if result != 0 && !hidden {
			resultReceipt := receipt("friend", row.ID, strconv.FormatUint(row.UserID, 10), "result", normalizeRead(row.SenderRead), 0, result, row.ReqTime, row.HandledAt, now)
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&resultReceipt).Error; err != nil {
				return fmt.Errorf("backfill friend result receipt: %w", err)
			}
		}
	}
	type groupRow struct {
		ID, GroupID   uint64
		ReqID         string
		HandleResult  *int
		ReceiverRead  int
		ReqTime       *time.Time
		HandleTime    *time.Time
		InvalidReason string
		SourceType    int
	}
	var groups []groupRow
	if err := db.Table("group_requests").Find(&groups).Error; err != nil {
		return fmt.Errorf("read group requests for receipts: %w", err)
	}
	for _, row := range groups {
		created := now
		if row.ReqTime != nil {
			created = *row.ReqTime
		}
		result, _ := legacyStatus(row.HandleResult)
		hidden := result == 3 || row.SourceType == 2 || row.InvalidReason == "migration_duplicate_pending" || row.InvalidReason == "migration_unconfirmed_invitation"
		isRead := normalizeRead(row.ReceiverRead)
		if hidden {
			isRead = 1
		}
		var admins []uint64
		if err := db.Table("group_members").Where("group_id = ? AND role_level IN ?", row.GroupID, []int{1, 2}).Pluck("user_id", &admins).Error; err != nil {
			return fmt.Errorf("list group administrators: %w", err)
		}
		for _, admin := range admins {
			apply := receipt("group", row.ID, strconv.FormatUint(admin, 10), "apply", isRead, boolInt(result == 0), result, created, row.HandleTime, now)
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&apply).Error; err != nil {
				return fmt.Errorf("backfill group apply receipt: %w", err)
			}
		}
		if result != 0 && !hidden {
			resultReceipt := receipt("group", row.ID, row.ReqID, "result", 0, 0, result, created, row.HandleTime, now)
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&resultReceipt).Error; err != nil {
				return fmt.Errorf("backfill group result receipt: %w", err)
			}
		}
	}
	return nil
}

func receipt(requestType string, requestID uint64, receiver, kind string, isRead, actionable, result int, created time.Time, resolved *time.Time, now time.Time) objects.SocialRequestReceipt {
	isRead = normalizeRead(isRead)
	var readAt *time.Time
	if isRead == 1 {
		readAt = &now
	}
	if result == 0 {
		resolved = nil
	} else if resolved == nil {
		resolved = &now
	}
	return objects.SocialRequestReceipt{RequestType: requestType, RequestID: requestID, ReceiverID: receiver, ReceiptKind: kind, IsRead: isRead, IsActionable: normalizeRead(actionable), Result: result, CreatedAt: created, ReadAt: readAt, ResolvedAt: resolved}
}

func normalizeRead(value int) int {
	if value == 0 {
		return 0
	}
	return 1
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func createAndValidateIndexes(db *gorm.DB) error {
	indexes := []struct {
		model any
		name  string
	}{
		{&objects.Friend{}, "uk_friends_user_friend"}, {&objects.FriendRequest{}, "uk_friend_requests_active_key"},
		{&objects.GroupRequest{}, "uk_group_requests_active_key"}, {&objects.GroupRequest{}, "uk_group_requests_source_invitation"},
		{&objects.GroupRequest{}, "idx_group_request_lookup"},
	}
	for _, index := range indexes {
		if !db.Migrator().HasIndex(index.model, index.name) {
			if err := db.Migrator().CreateIndex(index.model, index.name); err != nil {
				return fmt.Errorf("create index %s: %w", index.name, err)
			}
		}
	}
	if !db.Migrator().HasIndex(&objects.GroupMember{}, "uk_member") {
		if err := db.Migrator().CreateIndex(&objects.GroupMember{}, "uk_member"); err != nil {
			return fmt.Errorf("create group member unique index: %w", err)
		}
	}
	return validateSchema(db)
}

func validateSchema(db *gorm.DB) error {
	for _, table := range []string{"group_invitations", "social_request_receipts", "social_notification_outbox"} {
		if !db.Migrator().HasTable(table) {
			return fmt.Errorf("required table %s is missing", table)
		}
	}
	wanted := []struct {
		table   string
		columns []string
	}{
		{table: "friends", columns: []string{"user_id", "friend_uid"}},
		{table: "friend_requests", columns: []string{"active_key"}},
		{table: "group_requests", columns: []string{"active_key"}},
		{table: "group_requests", columns: []string{"source_invitation_id"}},
		{table: "group_members", columns: []string{"group_id", "user_id"}},
	}
	for _, expected := range wanted {
		indexes, err := db.Migrator().GetIndexes(expected.table)
		if err != nil {
			return fmt.Errorf("inspect %s indexes: %w", expected.table, err)
		}
		found := false
		for _, index := range indexes {
			unique, ok := index.Unique()
			if ok && unique && equalStrings(index.Columns(), expected.columns) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s missing unique index on %v", expected.table, expected.columns)
		}
	}
	return nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
