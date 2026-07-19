package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func TestMigrateLegacyDatabase(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		dsn    func(*testing.T) string
	}{
		{name: "sqlite", driver: "sqlite", dsn: func(t *testing.T) string { return t.TempDir() + "/social.db" }},
		{name: "mysql", driver: "mysql", dsn: envDSN("SOCIAL_MIGRATION_TEST_MYSQL_DSN")},
		{name: "postgres", driver: "postgres", dsn: envDSN("SOCIAL_MIGRATION_TEST_POSTGRES_DSN")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := test.dsn(t)
			if dsn == "" {
				t.Skip("integration DSN is not set")
			}
			db := openTestDB(t, test.driver, dsn)
			dropTestTables(t, db)
			require.NoError(t, createLegacySchema(db))
			require.NoError(t, seedLegacySchema(db))

			now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
			report, err := Migrate(context.Background(), db, test.driver, now)
			require.NoError(t, err)
			require.Len(t, report.FriendRequestDuplicates, 1)
			require.Equal(t, uint64(10), report.FriendRequestDuplicates[0].KeepID)
			require.Equal(t, []uint64{11}, report.FriendRequestDuplicates[0].ChangedIDs)
			require.Len(t, report.GroupRequestDuplicates, 1)
			require.Equal(t, uint64(20), report.GroupRequestDuplicates[0].KeepID)
			require.Equal(t, []uint64{21}, report.GroupRequestDuplicates[0].ChangedIDs)
			require.Len(t, report.FriendMerges, 1)
			require.Equal(t, []uint64{2}, report.FriendMerges[0].DeleteIDs)
			require.GreaterOrEqual(t, len(report.FriendMerges[0].Conflicts), 7)
			require.Len(t, report.RecoverableInvitations, 2)
			require.Len(t, report.NonUserInvitations, 1)
			require.Empty(t, report.AmbiguousInvitations)
			require.NotEmpty(t, report.OneWayFriends)
			require.Len(t, report.OneWayPendingRequests, 1)
			require.Equal(t, []uint64{12}, report.ResolvedFriendRequestIDs)
			require.Equal(t, []uint64{22}, report.ResolvedGroupRequestIDs)

			assertMigratedData(t, db)
			assertConstraints(t, db)

			repeated, err := Migrate(context.Background(), db, test.driver, now.Add(time.Hour))
			require.NoError(t, err)
			require.True(t, repeated.AlreadyApplied)
			var versions int64
			require.NoError(t, db.Model(&migrationRecord{}).Where("version = ?", migrationVersion).Count(&versions).Error)
			require.Equal(t, int64(1), versions)
		})
	}
}

func TestRepairInvitationReceiptResults(t *testing.T) {
	db := openTestDB(t, "sqlite", t.TempDir()+"/receipt-fix.db")
	require.NoError(t, db.AutoMigrate(&migrationRecord{}, &objects.GroupInvitation{}, &objects.SocialRequestReceipt{}))
	now := time.Now().UTC()
	invitations := []objects.GroupInvitation{
		{ID: 1, GroupID: 1, InviterUID: 1, InviteeUID: 2, Status: 3, CreatedAt: now, ExpiresAt: now},
		{ID: 2, GroupID: 1, InviterUID: 1, InviteeUID: 3, Status: 4, CreatedAt: now, ExpiresAt: now},
	}
	require.NoError(t, db.Create(&invitations).Error)
	receipts := []objects.SocialRequestReceipt{
		{RequestType: "group_invite", RequestID: 1, ReceiverID: "2", ReceiptKind: "invite", Result: 3, CreatedAt: now},
		{RequestType: "group_invite", RequestID: 2, ReceiverID: "3", ReceiptKind: "invite", Result: 4, CreatedAt: now},
	}
	require.NoError(t, db.Create(&receipts).Error)
	require.NoError(t, repairInvitationReceiptResults(db, now))

	var repaired []objects.SocialRequestReceipt
	require.NoError(t, db.Order("request_id").Find(&repaired).Error)
	require.Equal(t, 4, repaired[0].Result)
	require.Equal(t, 3, repaired[1].Result)
	require.NoError(t, repairInvitationReceiptResults(db, now.Add(time.Hour)))
	var records int64
	require.NoError(t, db.Model(&migrationRecord{}).Where("version = ?", receiptResultFixVersion).Count(&records).Error)
	require.Equal(t, int64(1), records)
}

func TestMigrateInvitationStatusOrderAndReceipts(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		dsn    func(*testing.T) string
	}{
		{name: "sqlite", driver: "sqlite", dsn: func(t *testing.T) string { return t.TempDir() + "/status-swap.db" }},
		{name: "mysql", driver: "mysql", dsn: envDSN("SOCIAL_MIGRATION_TEST_MYSQL_DSN")},
		{name: "postgres", driver: "postgres", dsn: envDSN("SOCIAL_MIGRATION_TEST_POSTGRES_DSN")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := test.dsn(t)
			if dsn == "" {
				t.Skip("integration DSN is not set")
			}
			db := openTestDB(t, test.driver, dsn)
			dropTestTables(t, db)
			now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
			require.NoError(t, createLegacySchema(db))
			_, err := Migrate(context.Background(), db, test.driver, now.Add(-time.Hour))
			require.NoError(t, err)
			require.NoError(t, db.Where("version IN ?", []string{invitationStatusSwapVersion, invitationReceiptCanonicalVersion}).Delete(&migrationRecord{}).Error)
			require.NoError(t, db.Clauses(clause.OnConflict{DoNothing: true}).Create(&migrationRecord{Version: receiptResultFixVersion, Description: "existing receipt repair", AppliedAt: now.Add(-time.Hour)}).Error)
			require.NoError(t, db.Create(&[]objects.GroupInvitation{
				{ID: 1, GroupID: 10, InviterUID: 1, InviteeUID: 2, Status: 3, CreatedAt: now, ExpiresAt: now}, // old expired
				{ID: 2, GroupID: 10, InviterUID: 1, InviteeUID: 3, Status: 4, CreatedAt: now, ExpiresAt: now}, // old invalidated
			}).Error)
			require.NoError(t, db.Create(&[]objects.SocialRequestReceipt{
				{RequestType: "group_invite", RequestID: 1, ReceiverID: "2", ReceiptKind: "invite", Result: 4, CreatedAt: now},
				{RequestType: "group_invite", RequestID: 2, ReceiverID: "3", ReceiptKind: "invite", Result: 3, CreatedAt: now},
			}).Error)

			report, err := Migrate(context.Background(), db, test.driver, now)
			require.NoError(t, err)
			require.True(t, report.AlreadyApplied)

			var invitations []objects.GroupInvitation
			require.NoError(t, db.Order("id").Find(&invitations).Error)
			require.Equal(t, []int{4, 3}, []int{invitations[0].Status, invitations[1].Status})
			var receipts []objects.SocialRequestReceipt
			require.NoError(t, db.Order("request_id").Find(&receipts).Error)
			require.Equal(t, []int{4, 3}, []int{receipts[0].Result, receipts[1].Result})

			repeated, err := Migrate(context.Background(), db, test.driver, now.Add(time.Hour))
			require.NoError(t, err)
			require.True(t, repeated.AlreadyApplied)
			require.NoError(t, db.Order("id").Find(&invitations).Error)
			require.Equal(t, []int{4, 3}, []int{invitations[0].Status, invitations[1].Status})
			require.NoError(t, db.Order("request_id").Find(&receipts).Error)
			require.Equal(t, []int{4, 3}, []int{receipts[0].Result, receipts[1].Result})
			var versions int64
			require.NoError(t, db.Model(&migrationRecord{}).Where("version IN ?", []string{invitationStatusSwapVersion, invitationReceiptCanonicalVersion}).Count(&versions).Error)
			require.Equal(t, int64(2), versions)
		})
	}
}

func TestMigrateInvitationStatusOrderAuditRollsBack(t *testing.T) {
	db := openTestDB(t, "sqlite", t.TempDir()+"/status-audit.db")
	require.NoError(t, db.AutoMigrate(&migrationRecord{}, &objects.GroupInvitation{}))
	now := time.Now().UTC()
	require.NoError(t, db.Create(&objects.GroupInvitation{ID: 1, GroupID: 1, InviterUID: 1, InviteeUID: 2, Status: 6, CreatedAt: now, ExpiresAt: now}).Error)

	err := migrateInvitationStatusOrder(db, now, true)
	require.ErrorContains(t, err, "outside 0..4")
	var versions int64
	require.NoError(t, db.Model(&migrationRecord{}).Where("version = ?", invitationStatusSwapVersion).Count(&versions).Error)
	require.Zero(t, versions)
	var invitation objects.GroupInvitation
	require.NoError(t, db.First(&invitation, 1).Error)
	require.Equal(t, 6, invitation.Status)
}

func TestCanonicalInvitationReceiptRepairRequiresStatusMigration(t *testing.T) {
	db := openTestDB(t, "sqlite", t.TempDir()+"/receipt-order.db")
	require.NoError(t, db.AutoMigrate(&migrationRecord{}, &objects.GroupInvitation{}, &objects.SocialRequestReceipt{}))
	err := repairCanonicalInvitationReceiptResults(db, time.Now().UTC())
	require.ErrorContains(t, err, "requires status migration")
	var versions int64
	require.NoError(t, db.Model(&migrationRecord{}).Where("version = ?", invitationReceiptCanonicalVersion).Count(&versions).Error)
	require.Zero(t, versions)
}

func TestMigrateBlocksAmbiguousInvitation(t *testing.T) {
	db := openTestDB(t, "sqlite", t.TempDir()+"/ambiguous.db")
	require.NoError(t, createLegacySchema(db))
	require.NoError(t, db.Exec(`INSERT INTO group_requests(id,req_id,group_id,req_time,join_source,handle_result,receiver_read) VALUES (1,'20',10,CURRENT_TIMESTAMP,2,0,0)`).Error)
	report, err := Migrate(context.Background(), db, "sqlite", time.Now().UTC())
	require.ErrorContains(t, err, "manual review")
	require.Len(t, report.AmbiguousInvitations, 1)
	var versions int64
	require.NoError(t, db.Model(&migrationRecord{}).Where("version = ?", migrationVersion).Count(&versions).Error)
	require.Zero(t, versions)
}

func TestMigrateBlocksUnsafeLegacyData(t *testing.T) {
	tests := []struct {
		name string
		seed string
		want string
	}{
		{name: "unknown invitation status", seed: `INSERT INTO group_requests(id,req_id,group_id,req_time,join_source,inviter_user_id,handle_result,receiver_read) VALUES (1,'20',10,CURRENT_TIMESTAMP,2,30,4,0)`, want: "manual review"},
		{name: "non numeric invitee", seed: `INSERT INTO group_requests(id,req_id,group_id,req_time,join_source,inviter_user_id,handle_result,receiver_read) VALUES (1,'not-a-uid',10,CURRENT_TIMESTAMP,2,30,0,0)`, want: "manual review"},
		{name: "invalid friend tags", seed: `INSERT INTO friends(id,user_id,friend_uid,friend_tags,created_at) VALUES (1,1,2,'[',CURRENT_TIMESTAMP),(2,1,2,'[]',CURRENT_TIMESTAMP)`, want: "invalid friend_tags JSON"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := openTestDB(t, "sqlite", t.TempDir()+"/unsafe.db")
			require.NoError(t, createLegacySchema(db))
			require.NoError(t, db.Exec(test.seed).Error)
			_, err := Migrate(context.Background(), db, "sqlite", time.Now().UTC())
			require.ErrorContains(t, err, test.want)
			var versions int64
			require.NoError(t, db.Model(&migrationRecord{}).Where("version IN ?", []string{migrationVersion, migrationDataVersion}).Count(&versions).Error)
			require.Zero(t, versions)
			if test.name == "invalid friend tags" {
				var rows int64
				require.NoError(t, db.Table("friends").Where("user_id = ? AND friend_uid = ?", 1, 2).Count(&rows).Error)
				require.Equal(t, int64(2), rows)
			}
		})
	}
}

func TestMigrateValidatesInvitationEntitiesWhenTablesExist(t *testing.T) {
	db := openTestDB(t, "sqlite", t.TempDir()+"/entities.db")
	require.NoError(t, createLegacySchema(db))
	require.NoError(t, db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE groups (id INTEGER PRIMARY KEY)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO users(id) VALUES (20),(30)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO group_requests(id,req_id,group_id,req_time,join_source,inviter_user_id,handle_result,receiver_read) VALUES (1,'20',10,CURRENT_TIMESTAMP,2,30,0,0)`).Error)
	report, err := Migrate(context.Background(), db, "sqlite", time.Now().UTC())
	require.ErrorContains(t, err, "manual review")
	require.Len(t, report.AmbiguousInvitations, 1)
	require.Equal(t, "group does not exist", report.AmbiguousInvitations[0].Reason)
}

func TestMigrateResumesAfterCommittedDataPhase(t *testing.T) {
	db := openTestDB(t, "sqlite", t.TempDir()+"/resume.db")
	require.NoError(t, createLegacySchema(db))
	require.NoError(t, seedLegacySchema(db))
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.AutoMigrate(&migrationRecord{}))
	require.NoError(t, ensureSchema(db))
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		report := newReport()
		if err := migrateData(tx, "sqlite", now, &report); err != nil {
			return err
		}
		return tx.Create(&migrationRecord{Version: migrationDataVersion, Description: "test data phase", AppliedAt: now}).Error
	}))
	require.False(t, db.Migrator().HasIndex(&objects.Friend{}, "uk_friends_user_friend"))

	report, err := Migrate(context.Background(), db, "sqlite", now.Add(time.Hour))
	require.NoError(t, err)
	require.True(t, report.DataAlreadyApplied)
	require.Empty(t, report.FriendMerges)
	require.True(t, db.Migrator().HasIndex(&objects.Friend{}, "uk_friends_user_friend"))
	var versions int64
	require.NoError(t, db.Model(&migrationRecord{}).Where("version = ?", migrationVersion).Count(&versions).Error)
	require.Equal(t, int64(1), versions)
}

func envDSN(key string) func(*testing.T) string {
	return func(*testing.T) string { return os.Getenv(key) }
}

func openTestDB(t *testing.T, driver, dsn string) *gorm.DB {
	t.Helper()
	var dialector gorm.Dialector
	switch driver {
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		t.Fatalf("unsupported test driver %s", driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	return db
}

func dropTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, table := range []string{"social_notification_outbox", "social_request_receipts", "group_invitations", "schema_migrations", "group_requests", "friend_requests", "group_members", "friends"} {
		require.NoError(t, db.Migrator().DropTable(table))
	}
}

func createLegacySchema(db *gorm.DB) error {
	statements := []string{
		`CREATE TABLE friends (id INTEGER PRIMARY KEY, user_id BIGINT NOT NULL, friend_uid BIGINT NOT NULL, remark VARCHAR(255), add_source INTEGER, blacklisted INTEGER NOT NULL DEFAULT 0, moments_permission INTEGER NOT NULL DEFAULT 0, notify_enabled INTEGER NOT NULL DEFAULT 1, pinned INTEGER NOT NULL DEFAULT 0, muted INTEGER NOT NULL DEFAULT 0, friend_tags TEXT, created_at TIMESTAMP NULL)`,
		`CREATE TABLE friend_requests (id INTEGER PRIMARY KEY, user_id BIGINT NOT NULL, req_uid BIGINT NOT NULL, req_msg VARCHAR(255), req_time TIMESTAMP NOT NULL, handle_result INTEGER, handle_msg VARCHAR(255), handled_at TIMESTAMP NULL, status INTEGER, read_state INTEGER NOT NULL DEFAULT 0, receiver_read INTEGER NOT NULL DEFAULT 0, sender_read INTEGER NOT NULL DEFAULT 0, remark VARCHAR(64) NOT NULL DEFAULT '')`,
		`CREATE TABLE group_requests (id INTEGER PRIMARY KEY, req_id VARCHAR(64) NOT NULL, group_id BIGINT NOT NULL, req_msg VARCHAR(255), req_time TIMESTAMP NULL, join_source INTEGER, inviter_user_id BIGINT, handle_user_id BIGINT, handle_time TIMESTAMP NULL, handle_result INTEGER, receiver_read INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE group_members (id INTEGER PRIMARY KEY, group_id BIGINT NOT NULL, user_id BIGINT NOT NULL, role_level INTEGER NOT NULL, join_time TIMESTAMP NULL, join_source INTEGER, inviter_uid BIGINT, operator_uid BIGINT, group_nickname VARCHAR(64) NOT NULL DEFAULT '', group_remark VARCHAR(255) NOT NULL DEFAULT '')`,
		`CREATE UNIQUE INDEX uk_member ON group_members(group_id,user_id)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedLegacySchema(db *gorm.DB) error {
	statements := []string{
		`INSERT INTO friends(id,user_id,friend_uid,remark,add_source,blacklisted,moments_permission,notify_enabled,pinned,muted,friend_tags,created_at) VALUES (1,1,2,'first',1,0,1,1,0,0,'["a","shared"]','2026-01-01'),(2,1,2,'conflict',2,1,2,0,1,1,'["shared","b"]','2026-01-02'),(3,2,1,'reverse',NULL,0,0,1,0,0,'[]','2026-01-01'),(4,8,9,'one-way',NULL,0,0,1,0,0,'[]','2026-01-01')`,
		`INSERT INTO friend_requests(id,user_id,req_uid,req_time,handle_result,status,receiver_read,sender_read) VALUES (10,3,4,'2026-07-02',0,1,9,0),(11,3,4,'2026-07-01',0,1,0,0),(12,1,2,'2026-07-03',0,1,1,0),(13,5,6,'2026-07-04',2,1,1,0),(14,8,9,'2026-07-05',0,1,0,0)`,
		`INSERT INTO group_members(id,group_id,user_id,role_level) VALUES (1,100,50,2),(2,100,51,1),(3,100,22,0)`,
		`INSERT INTO group_requests(id,req_id,group_id,req_msg,req_time,join_source,inviter_user_id,handle_result,receiver_read) VALUES (20,'20',100,'new','2026-07-02',1,NULL,0,1),(21,'20',100,'old','2026-07-01',1,NULL,0,0),(22,'22',100,'already member','2026-07-03',1,NULL,0,0),(23,'23',100,'member invite','2026-07-16',2,50,0,0),(24,'24',100,'link flow','2026-07-10',3,50,1,1),(25,'22',100,'invite already member','2026-07-16',2,50,0,0)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func assertMigratedData(t *testing.T, db *gorm.DB) {
	t.Helper()
	var friendRows []struct {
		ID           uint64
		HandleResult int
		ActiveKey    *string
	}
	require.NoError(t, db.Table("friend_requests").Where("user_id = ? AND req_uid = ?", 3, 4).Order("id").Scan(&friendRows).Error)
	require.Equal(t, 0, friendRows[0].HandleResult)
	require.Equal(t, "friend:3:4", *friendRows[0].ActiveKey)
	require.Equal(t, 3, friendRows[1].HandleResult)
	require.Nil(t, friendRows[1].ActiveKey)

	var groupRows []struct {
		ID            uint64
		HandleResult  int
		ActiveKey     *string
		InvalidReason string
	}
	require.NoError(t, db.Table("group_requests").Where("req_id = ?", "20").Order("id").Scan(&groupRows).Error)
	require.Equal(t, 0, groupRows[0].HandleResult)
	require.Equal(t, "group:direct:100:20", *groupRows[0].ActiveKey)
	require.Equal(t, 3, groupRows[1].HandleResult)
	require.Equal(t, "migration_duplicate_pending", groupRows[1].InvalidReason)

	var invitation objects.GroupInvitation
	require.NoError(t, db.First(&invitation, 23).Error)
	require.Equal(t, uint64(50), invitation.InviterUID)
	require.Equal(t, uint64(23), invitation.InviteeUID)
	require.Equal(t, 0, invitation.Status)
	var invitations int64
	require.NoError(t, db.Model(&objects.GroupInvitation{}).Count(&invitations).Error)
	require.Equal(t, int64(2), invitations)
	var memberInvitation objects.GroupInvitation
	require.NoError(t, db.First(&memberInvitation, 25).Error)
	require.Equal(t, 1, memberInvitation.Status)
	require.NotNil(t, memberInvitation.HandledAt)

	var kept struct{ Remark, FriendTags string }
	require.NoError(t, db.Table("friends").Where("user_id = ? AND friend_uid = ?", 1, 2).First(&kept).Error)
	require.Equal(t, "first", kept.Remark)
	require.Equal(t, `["a","shared","b"]`, kept.FriendTags)
	var settings struct {
		Blacklisted                  bool
		MomentsPermission            int
		NotifyEnabled, Pinned, Muted bool
		AddSource                    *int
	}
	require.NoError(t, db.Table("friends").Where("user_id = ? AND friend_uid = ?", 1, 2).First(&settings).Error)
	require.True(t, settings.Blacklisted)
	require.Equal(t, 2, settings.MomentsPermission)
	require.False(t, settings.NotifyEnabled)
	require.True(t, settings.Pinned)
	require.True(t, settings.Muted)
	require.Equal(t, 1, *settings.AddSource)
	for _, table := range []string{"group_invitations", "social_request_receipts", "social_notification_outbox"} {
		require.True(t, db.Migrator().HasTable(table))
	}
	var receipts int64
	require.NoError(t, db.Model(&objects.SocialRequestReceipt{}).Count(&receipts).Error)
	require.Greater(t, receipts, int64(0))
	var invitationReceipts int64
	require.NoError(t, db.Model(&objects.SocialRequestReceipt{}).Where("request_type = ? AND request_id = ? AND receipt_kind = ?", "group_invite", 23, "invite").Count(&invitationReceipts).Error)
	require.Equal(t, int64(1), invitationReceipts)
	var legacyInviteRequest struct {
		HandleResult       int
		SourceInvitationID *uint64
		InvalidReason      string
	}
	require.NoError(t, db.Table("group_requests").Where("id = ?", 23).First(&legacyInviteRequest).Error)
	require.Equal(t, 3, legacyInviteRequest.HandleResult)
	require.Nil(t, legacyInviteRequest.SourceInvitationID)
	require.Equal(t, "migration_unconfirmed_invitation", legacyInviteRequest.InvalidReason)
	var memberInviteRequest struct {
		HandleResult       int
		SourceInvitationID *uint64
		ActualJoinSource   *int
	}
	require.NoError(t, db.Table("group_requests").Where("id = ?", 25).First(&memberInviteRequest).Error)
	require.Equal(t, 1, memberInviteRequest.HandleResult)
	require.Nil(t, memberInviteRequest.SourceInvitationID)
	require.Equal(t, 2, *memberInviteRequest.ActualJoinSource)
	var hiddenResults int64
	require.NoError(t, db.Model(&objects.SocialRequestReceipt{}).Where("receipt_kind = ? AND request_id IN ?", "result", []uint64{11, 21, 23}).Count(&hiddenResults).Error)
	require.Zero(t, hiddenResults)
	var terminalReceipts []objects.SocialRequestReceipt
	require.NoError(t, db.Where("result <> 0").Find(&terminalReceipts).Error)
	for _, receipt := range terminalReceipts {
		require.NotNil(t, receipt.ResolvedAt)
		if receipt.IsRead == 1 {
			require.NotNil(t, receipt.ReadAt)
			require.Equal(t, time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC), receipt.ReadAt.UTC())
		}
	}
}

func assertConstraints(t *testing.T, db *gorm.DB) {
	t.Helper()
	assertInsertFails(t, db, `INSERT INTO friends(id,user_id,friend_uid,remark,blacklisted,moments_permission,notify_enabled,pinned,muted,friend_tags) VALUES (1001,1,2,'',0,0,1,0,0,'')`)
	assertInsertFails(t, db, `INSERT INTO friend_requests(id,user_id,req_uid,req_time,handle_result,status,read_state,receiver_read,sender_read,remark,active_key) VALUES (1001,30,31,CURRENT_TIMESTAMP,0,1,0,0,0,'','friend:3:4')`)
	require.NoError(t, db.Exec(`INSERT INTO friend_requests(id,user_id,req_uid,req_time,handle_result,status,read_state,receiver_read,sender_read,remark,active_key) VALUES (1002,40,41,CURRENT_TIMESTAMP,1,1,0,0,0,'',NULL),(1003,42,43,CURRENT_TIMESTAMP,2,1,0,0,0,'',NULL)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO group_requests(id,req_id,group_id,receiver_read,source_type,invalid_reason,source_invitation_id) VALUES (1001,'70',100,0,1,'',NULL),(1002,'71',100,0,1,'',NULL),(1003,'72',100,0,2,'',9001)`).Error)
	assertInsertFails(t, db, `INSERT INTO group_requests(id,req_id,group_id,receiver_read,source_type,invalid_reason,source_invitation_id) VALUES (1004,'73',100,0,2,'',9001)`)
	for _, pair := range []struct{ table, index string }{{"group_invitations", "idx_group_invitation_expiry"}, {"social_request_receipts", "uk_social_request_receipt"}, {"social_notification_outbox", "uk_social_notification"}} {
		require.True(t, db.Migrator().HasIndex(pair.table, pair.index), fmt.Sprintf("missing %s", pair.index))
	}
}

func assertInsertFails(t *testing.T, db *gorm.DB, statement string) {
	t.Helper()
	require.Error(t, db.Exec(statement).Error)
}
