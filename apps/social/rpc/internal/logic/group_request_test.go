package logic

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newGroupTestContext(t *testing.T) (*svc.ServiceContext, *notifyRecorder) {
	t.Helper()
	database := openGroupTestDB(t)
	if database.Dialector.Name() == "sqlite" {
		createGroupSQLiteSchema(t, database)
	} else {
		require.NoError(t, database.AutoMigrate(
			&objects.Group{}, &objects.GroupMember{}, &objects.GroupRequest{},
			&objects.GroupInvitation{}, &objects.GroupInviteLink{}, &objects.RelationOutbox{}, &objects.SocialRequestReceipt{}, &objects.SocialNotificationOutbox{},
		))
	}
	recorder := &notifyRecorder{}
	users := make(map[string]*user.UserEntity)
	for i := 1; i <= 20; i++ {
		id := strconv.Itoa(i)
		users[id] = &user.UserEntity{Id: id, Nickname: "user-" + id, Status: 1}
	}
	users["20"].Status = 0
	return &svc.ServiceContext{DB: database, User: &userLookupStub{users: users}, CommonNotifyClient: recorder}, recorder
}

func TestGroupRequestListPaginationAndFields(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	group := testGroup(1, false)
	require.NoError(t, svcCtx.DB.Create(group).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	now := time.Now()
	for i, result := range []int{0, 1, 3} {
		id := uint64(math.MaxInt32) + uint64(i) + 1
		reqTime := now.Add(time.Duration(i) * time.Second)
		sourceInvitation := id + 100
		req := objects.GroupRequest{ID: id, ReqID: "3", GroupID: 1, ReqMsg: "join", ReqTime: &reqTime, JoinSource: intPtr(2), HandleResult: intPtr(result), SourceType: 2, SourceInvitationID: &sourceInvitation, HandleMsg: "handled", InvalidReason: "reason"}
		require.NoError(t, svcCtx.DB.Create(&req).Error)
		require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroup, id, "1", receiptKindApply, 1, result, now, true, nil))
	}
	statusFilter := int32(0)
	filtered, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{Ids: []string{"1"}, ActorUid: "1", Class: "2", Status: &statusFilter, Page: 1, Size: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), filtered.Total)
	require.Greater(t, filtered.List[0].RequestId, uint64(math.MaxInt32))
	require.Equal(t, "3", filtered.List[0].ApplicantUid)
	require.True(t, filtered.List[0].Actionable)
	require.Equal(t, int32(2), filtered.List[0].SourceType)
	all, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{Ids: []string{"1"}, ActorUid: "1", Class: "2", Page: 2, Size: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), all.Total)
	require.Len(t, all.List, 1)
}

func TestGroupRequestListSeparatesSentAndReceiptOwnedReceived(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	now := time.Now()
	own := objects.GroupRequest{ReqID: "1", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
	other := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&own).Error)
	require.NoError(t, svcCtx.DB.Create(&other).Error)

	sent, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{Ids: []string{"1"}, ActorUid: "1", Class: "1"})
	require.NoError(t, err)
	require.Equal(t, int64(1), sent.Total)
	require.Equal(t, own.ID, sent.List[0].RequestId)

	received, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{Ids: []string{"1"}, ActorUid: "1", Class: "2"})
	require.NoError(t, err)
	require.Zero(t, received.Total)
	require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroup, other.ID, "1", receiptKindApply, 0, receiptPending, now, false, nil))
	received, err = NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{Ids: []string{"1"}, ActorUid: "1", Class: "2"})
	require.NoError(t, err)
	require.Equal(t, int64(1), received.Total)
	require.False(t, received.List[0].Actionable)
}

func TestGroupRequestSentProjectionRedactsHandler(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	now := time.Now()
	handler := uint64(2)
	request := objects.GroupRequest{ReqID: "1", GroupID: 1, ReqTime: &now, HandleResult: intPtr(1), HandleUserID: &handler, SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&request).Error)

	resp, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{
		ActorUid: "1", Ids: []string{"1"}, Class: "1",
	})
	require.NoError(t, err)
	require.Len(t, resp.List, 1)
	require.Empty(t, resp.List[0].HandleUid)
}

func TestGetGroupPutListByUidRepeatedActorCompatibility(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	now := time.Now()
	require.NoError(t, svcCtx.DB.Create(&objects.GroupRequest{ReqID: "1", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}).Error)

	resp, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{
		ActorUid: "1", Ids: []string{"1", "1"}, Class: "1",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Total)

	_, err = NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{
		ActorUid: "1", Ids: []string{"1", "2"}, Class: "1",
	})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestGroupScopedActorValidationAndForgery(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	for _, tc := range []struct {
		name, actor, legacy string
		code                codes.Code
	}{
		{name: "missing", legacy: "1", code: codes.Unauthenticated},
		{name: "malformed", actor: "invalid", legacy: "invalid", code: codes.InvalidArgument},
		{name: "mismatch", actor: "1", legacy: "2", code: codes.PermissionDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewGroupPutinListLogic(context.Background(), svcCtx).GroupPutinList(&social.GroupPutinListReq{ActorUid: tc.actor, UserId: tc.legacy, Class: 1})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewMarkGroupReqReadLogic(context.Background(), svcCtx).MarkGroupReqRead(&social.MarkGroupReqReadReq{ActorUid: tc.actor, UserId: tc.legacy, RequestIds: []uint64{1}})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewGroupRequestMessageCountLogic(context.Background(), svcCtx).GroupRequestMessageCount(&social.GroupRequestMessageCountReq{ActorUid: tc.actor, UserId: tc.legacy})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{ActorUid: tc.actor, UserId: tc.legacy, GroupId: "1"})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewGroupInviteLinkCreateLogic(context.Background(), svcCtx).GroupInviteLinkCreate(&social.GroupInviteLinkCreateReq{ActorUid: tc.actor, UserId: tc.legacy, GroupId: "1"})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewGroupInviteLinkListLogic(context.Background(), svcCtx).GroupInviteLinkList(&social.GroupInviteLinkListReq{ActorUid: tc.actor, UserId: tc.legacy, GroupId: "1"})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewGroupInviteLinkRevokeLogic(context.Background(), svcCtx).GroupInviteLinkRevoke(&social.GroupInviteLinkRevokeReq{ActorUid: tc.actor, UserId: tc.legacy, GroupId: "1", Token: "token"})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewGroupJoinByTokenLogic(context.Background(), svcCtx).GroupJoinByToken(&social.GroupJoinByTokenReq{ActorUid: tc.actor, UserId: tc.legacy, Token: "token"})
			require.Equal(t, tc.code, status.Code(err))
		})
	}

	_, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{ActorUid: "1", Ids: []string{"2"}, Class: "1"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestGroupJoinByTokenRequiresNormalActor(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, false)).Error)
	seedInviteLink(t, svcCtx.DB, "disabled", 1, 1, 1)
	_, err := NewGroupJoinByTokenLogic(context.Background(), svcCtx).GroupJoinByToken(&social.GroupJoinByTokenReq{
		ActorUid: "20", UserId: "20", Token: "disabled",
	})
	require.Equal(t, codes.NotFound, status.Code(err), err)
}

func TestGroupJoinByTokenAtomicSuccessAccounting(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, false)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedInviteLink(t, svcCtx.DB, "success", 1, 1, 1)

	resp, err := NewGroupJoinByTokenLogic(context.Background(), svcCtx).GroupJoinByToken(&social.GroupJoinByTokenReq{
		ActorUid: "3", UserId: "3", Token: "success",
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), resp.IsPass)
	require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupRequest{}))
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))
	var link objects.GroupInviteLink
	require.NoError(t, svcCtx.DB.Where("token = ?", "success").First(&link).Error)
	require.Equal(t, uint(1), link.UsedCount)
}

func TestGroupJoinByTokenAccountingFailureRollsBackApplication(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger injects usage accounting failure")
	}
	require.NoError(t, svcCtx.DB.Create(testGroup(1, false)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedInviteLink(t, svcCtx.DB, "rollback", 1, 1, 1)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_token_usage BEFORE UPDATE OF used_count ON group_invite_links BEGIN SELECT RAISE(ABORT, 'usage failure'); END`).Error)

	_, err := NewGroupJoinByTokenLogic(context.Background(), svcCtx).GroupJoinByToken(&social.GroupJoinByTokenReq{
		ActorUid: "3", UserId: "3", Token: "rollback",
	})
	require.Error(t, err)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupRequest{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))
	var link objects.GroupInviteLink
	require.NoError(t, svcCtx.DB.Where("token = ?", "rollback").First(&link).Error)
	require.Zero(t, link.UsedCount)
}

func TestGroupJoinByTokenCommitFailureRollsBackEverything(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite deferred foreign key injects commit failure")
	}
	sqlDB, err := svcCtx.DB.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, svcCtx.DB.Exec("PRAGMA foreign_keys = ON").Error)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TABLE token_commit_parent (id INTEGER PRIMARY KEY)`).Error)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TABLE token_commit_guard (parent_id INTEGER, FOREIGN KEY(parent_id) REFERENCES token_commit_parent(id) DEFERRABLE INITIALLY DEFERRED)`).Error)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_token_commit AFTER UPDATE OF used_count ON group_invite_links BEGIN INSERT INTO token_commit_guard(parent_id) VALUES (1); END`).Error)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, false)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedInviteLink(t, svcCtx.DB, "commit", 1, 1, 1)

	_, err = NewGroupJoinByTokenLogic(context.Background(), svcCtx).GroupJoinByToken(&social.GroupJoinByTokenReq{
		ActorUid: "3", UserId: "3", Token: "commit",
	})
	require.Error(t, err)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupRequest{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))
	var link objects.GroupInviteLink
	require.NoError(t, svcCtx.DB.Where("token = ?", "commit").First(&link).Error)
	require.Zero(t, link.UsedCount)
}

func TestGroupCreateEmitsCreatorMembershipEvent(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	resp, err := NewGroupCreateLogic(context.Background(), svcCtx).GroupCreate(&social.GroupCreateReq{Name: "new-group", CreatorUid: "1"})
	require.NoError(t, err)
	var memberCount, outboxCount int64
	require.NoError(t, svcCtx.DB.Model(&objects.GroupMember{}).Where("group_id = ? AND user_id = ?", resp.GroupId, 1).Count(&memberCount).Error)
	require.Equal(t, int64(1), memberCount)
	require.NoError(t, svcCtx.DB.Model(&objects.RelationOutbox{}).Where("event_type = ? AND group_id = ?", constants.RelationEventGroupMemberAdded, resp.GroupId).Count(&outboxCount).Error)
	require.Equal(t, int64(1), outboxCount)
}

func createGroupSQLiteSchema(t *testing.T, database *gorm.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, icon TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '', status INTEGER, creator_uid INTEGER NOT NULL,
			group_type INTEGER NOT NULL, is_verify INTEGER NOT NULL, notification TEXT,
			notification_uid INTEGER, created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE group_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
			role_level INTEGER NOT NULL, join_time DATETIME, join_source INTEGER, inviter_uid INTEGER,
			operator_uid INTEGER, group_nickname TEXT NOT NULL DEFAULT '', group_remark TEXT NOT NULL DEFAULT '',
			UNIQUE(group_id, user_id)
		)`,
		`CREATE TABLE group_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT, req_id TEXT NOT NULL, group_id INTEGER NOT NULL,
			req_msg TEXT, req_time DATETIME, join_source INTEGER, inviter_user_id INTEGER,
			handle_user_id INTEGER, handle_time DATETIME, handle_result INTEGER, receiver_read INTEGER NOT NULL DEFAULT 0,
			active_key TEXT UNIQUE, source_type INTEGER NOT NULL DEFAULT 1, source_invitation_id INTEGER UNIQUE,
			actual_join_source INTEGER, invalid_reason TEXT NOT NULL DEFAULT ''
			, handle_msg TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE group_invitations (
			id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER NOT NULL, inviter_uid INTEGER NOT NULL,
			invitee_uid INTEGER NOT NULL, inviter_role_snapshot INTEGER NOT NULL DEFAULT 0,
			message TEXT NOT NULL DEFAULT '', status INTEGER NOT NULL DEFAULT 0,
			reject_reason TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL,
			handled_at DATETIME, expires_at DATETIME NOT NULL
		)`,
		`CREATE TABLE group_invite_links (
			id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER NOT NULL, token TEXT NOT NULL UNIQUE,
			created_by INTEGER NOT NULL, expire_at DATETIME, max_uses INTEGER NOT NULL DEFAULT 0,
			used_count INTEGER NOT NULL DEFAULT 0, revoked INTEGER NOT NULL DEFAULT 0,
			revoked_at DATETIME, created_at DATETIME
		)`,
		`CREATE TABLE relation_outbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT, event_type TEXT NOT NULL, group_id TEXT NOT NULL DEFAULT '',
			payload TEXT NOT NULL, status INTEGER NOT NULL DEFAULT 0, created_at DATETIME, sent_at DATETIME
		)`,
		`CREATE TABLE social_request_receipts (
			id INTEGER PRIMARY KEY AUTOINCREMENT, request_type TEXT NOT NULL, request_id INTEGER NOT NULL,
			receiver_id TEXT NOT NULL, receipt_kind TEXT NOT NULL, is_read INTEGER NOT NULL DEFAULT 0,
			is_actionable INTEGER NOT NULL DEFAULT 0, result INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL, read_at DATETIME, resolved_at DATETIME,
			UNIQUE(request_type, request_id, receiver_id, receipt_kind)
		)`,
		`CREATE INDEX idx_group_invitation_invitee ON group_invitations(invitee_uid, status, created_at)`,
		`CREATE INDEX idx_group_invitation_group_invitee ON group_invitations(group_id, invitee_uid, status)`,
	}
	for _, statement := range statements {
		require.NoError(t, database.Exec(statement).Error)
	}
	require.NoError(t, database.AutoMigrate(&objects.SocialNotificationOutbox{}))
}

func openGroupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := "sqlite"
	dsn := "file:" + t.TempDir() + "/group.db?_busy_timeout=10000&_journal_mode=WAL"
	dialector := gorm.Dialector(sqlite.Open(dsn))
	if value := os.Getenv("SOCIAL_GROUP_TEST_MYSQL_DSN"); value != "" {
		name, dsn, dialector = "mysql", value, mysql.Open(value)
	} else if value := os.Getenv("SOCIAL_GROUP_TEST_POSTGRES_DSN"); value != "" {
		name, dsn, dialector = "postgres", value, postgres.Open(value)
	}
	database, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err, "%s DSN %s", name, dsn)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	if database.Dialector.Name() == "sqlite" {
		sqlDB.SetMaxOpenConns(20)
	}
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	return database
}

func TestGroupPutInValidationAndState(t *testing.T) {
	tests := []struct {
		name     string
		group    *objects.Group
		member   bool
		req      *social.GroupPutinReq
		code     codes.Code
		accepted bool
	}{
		{name: "missing actor", group: testGroup(1, true), req: groupPutInReq("", "1", "1"), code: codes.Unauthenticated},
		{name: "legacy actor mismatch", group: testGroup(1, true), req: groupPutInReq("3", "2", "1"), code: codes.PermissionDenied},
		{name: "missing group", req: groupPutInReq("3", "3", "99"), code: codes.NotFound},
		{name: "abnormal group", group: testAbnormalGroup(1), req: groupPutInReq("3", "3", "1"), code: codes.FailedPrecondition},
		{name: "already member", group: testGroup(1, true), member: true, req: groupPutInReq("3", "3", "1")},
		{name: "direct join", group: testGroup(1, false), req: groupPutInReq("3", "3", "1"), accepted: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, recorder := newGroupTestContext(t)
			if tt.group != nil {
				require.NoError(t, svcCtx.DB.Create(tt.group).Error)
			}
			if tt.member {
				seedGroupMember(t, svcCtx.DB, 1, 3, 0)
			}
			resp, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(tt.req)
			require.Equal(t, tt.code, status.Code(err))
			if tt.code != codes.OK {
				require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupRequest{}))
				return
			}
			if tt.member {
				require.True(t, resp.AlreadyMember)
				require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
			} else if tt.accepted {
				require.Equal(t, int32(1), resp.IsPass)
				require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
				require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
			}
			if tt.accepted {
				require.Zero(t, recorder.count())
				require.Zero(t, countRows(t, svcCtx.DB, &objects.SocialNotificationOutbox{}))
			} else {
				require.Zero(t, recorder.count())
			}
		})
	}
}

func TestGroupPutInIgnoresForgedPublicFieldsAndDeduplicates(t *testing.T) {
	svcCtx, recorder := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	req := &social.GroupPutinReq{
		ActorUid: "3", ReqId: "3", GroupId: "1", ReqMsg: "hello", ReqTime: 1,
		JoinSource: 99, InviterUid: "1",
	}
	first, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(req)
	require.NoError(t, err)
	second, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(req)
	require.NoError(t, err)
	require.Equal(t, first.RequestId, second.RequestId)
	require.True(t, second.AlreadyPending)
	var request objects.GroupRequest
	require.NoError(t, svcCtx.DB.First(&request, first.RequestId).Error)
	require.Equal(t, "3", request.ReqID)
	require.Equal(t, 1, request.SourceType)
	require.Equal(t, 1, *request.JoinSource)
	require.Nil(t, request.InviterUserID)
	require.Nil(t, request.HandleTime)
	require.Zero(t, recorder.count())
}

func TestGroupPutInConcurrentAndCooldown(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	const workers = 20
	var wg sync.WaitGroup
	ids := make(chan uint64, workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(groupPutInReq("3", "3", "1"))
			if err == nil {
				ids <- resp.RequestId
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var stable uint64
	for id := range ids {
		if stable == 0 {
			stable = id
		}
		require.Equal(t, stable, id)
	}
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupRequest{}))
	now := time.Now()
	require.NoError(t, svcCtx.DB.Model(&objects.GroupRequest{}).Where("id = ?", stable).
		Updates(map[string]any{"handle_result": 2, "handle_time": now, "active_key": nil}).Error)
	_, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(groupPutInReq("3", "3", "1"))
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	require.NoError(t, svcCtx.DB.Model(&objects.GroupRequest{}).Where("id = ?", stable).Update("req_time", now.Add(-time.Minute-time.Second)).Error)
	resp, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(groupPutInReq("3", "3", "1"))
	require.NoError(t, err)
	require.NotEqual(t, stable, resp.RequestId)
}

func TestGroupRequestReceiptsArePersonal(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 1)
	created, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(groupPutInReq("3", "3", "1"))
	require.NoError(t, err)

	var receipts []objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receipt_kind = ?", receiptTypeGroup, created.RequestId, receiptKindApply).Order("receiver_id").Find(&receipts).Error)
	require.Len(t, receipts, 2)
	require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.SocialNotificationOutbox{}))
	require.Equal(t, []string{"1", "2"}, []string{receipts[0].ReceiverID, receipts[1].ReceiverID})

	read, err := NewMarkGroupReqReadLogic(context.Background(), svcCtx).MarkGroupReqRead(&social.MarkGroupReqReadReq{UserId: "1", ActorUid: "1", RequestIds: []uint64{created.RequestId}})
	require.NoError(t, err)
	require.Equal(t, int32(0), read.Apply)
	otherCount, err := NewGroupRequestMessageCountLogic(context.Background(), svcCtx).GroupRequestMessageCount(&social.GroupRequestMessageCountReq{UserId: "2", ActorUid: "2"})
	require.NoError(t, err)
	require.Equal(t, int32(1), otherCount.Apply)

	_, err = NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{RequestId: created.RequestId, ActorUid: "1", HandleResult: 1})
	require.NoError(t, err)
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receipt_kind = ?", receiptTypeGroup, created.RequestId, receiptKindApply).Order("receiver_id").Find(&receipts).Error)
	require.Zero(t, receipts[0].IsActionable)
	require.Zero(t, receipts[1].IsActionable)
	require.Equal(t, 1, receipts[0].IsRead)
	require.Zero(t, receipts[1].IsRead)
	require.Equal(t, receiptAccepted, receipts[1].Result)
	require.Equal(t, int64(3), countRows(t, svcCtx.DB, &objects.SocialNotificationOutbox{}))

	var applicantResult objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ? AND receipt_kind = ?", receiptTypeGroup, created.RequestId, "3", receiptKindResult).First(&applicantResult).Error)
	require.Zero(t, applicantResult.IsRead)
	applicantCount, err := NewGroupRequestMessageCountLogic(context.Background(), svcCtx).GroupRequestMessageCount(&social.GroupRequestMessageCountReq{UserId: "3", ActorUid: "3"})
	require.NoError(t, err)
	require.Equal(t, int32(1), applicantCount.Result)
}

func TestGroupInvitationReceiptLifecycle(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	created, err := NewGroupInvitationCreateLogic(context.Background(), svcCtx).GroupInvitationCreate(&social.GroupInvitationCreateReq{ActorUid: "1", GroupId: "1", InviteeUid: "3"})
	require.NoError(t, err)
	count, err := NewGroupRequestMessageCountLogic(context.Background(), svcCtx).GroupRequestMessageCount(&social.GroupRequestMessageCountReq{UserId: "3", ActorUid: "3"})
	require.NoError(t, err)
	require.Equal(t, int32(1), count.Invite)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.SocialNotificationOutbox{}))

	read, err := NewGroupInvitationReadLogic(context.Background(), svcCtx).GroupInvitationRead(&social.GroupInvitationReadReq{ActorUid: "3", InvitationIds: []uint64{created.Invitation.Id}})
	require.NoError(t, err)
	require.Zero(t, read.Invite)
	_, err = NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{Id: created.Invitation.Id, ActorUid: "3", Result: 2})
	require.NoError(t, err)
	var receipt objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ?", receiptTypeGroupInvite, created.Invitation.Id).First(&receipt).Error)
	require.Zero(t, receipt.IsActionable)
	require.Equal(t, receiptRejected, receipt.Result)
}

func TestGroupInvitationListExpiresPendingBeforeFiltering(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	invitation := seedInvitation(t, svcCtx.DB, 1, 1, 3, 2)
	past := time.Now().Add(-time.Second)
	require.NoError(t, svcCtx.DB.Model(&objects.GroupInvitation{}).Where("id = ?", invitation.ID).Update("expires_at", past).Error)
	require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroupInvite, invitation.ID, "3", receiptKindInvite, 1, receiptPending, invitation.CreatedAt, false, nil))

	resp, err := NewGroupInvitationListLogic(context.Background(), svcCtx).GroupInvitationList(&social.GroupInvitationListReq{ActorUid: "3", Status: groupInvitationExpired})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Total)
	require.Equal(t, int32(groupInvitationExpired), resp.List[0].Status)
	require.False(t, resp.List[0].Actionable)
	var receipt objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ?", receiptTypeGroupInvite, invitation.ID).First(&receipt).Error)
	require.Equal(t, receiptExpired, receipt.Result)
}

func TestGroupAdminRoleChangesConvergePendingReceipts(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 0)
	now := time.Now()
	request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&request).Error)

	_, err := NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{UserId: "1", ActorUid: "1", GroupId: "1", MemberIds: []string{"2"}, IsAdmin: true})
	require.NoError(t, err)
	_, err = NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{UserId: "1", ActorUid: "1", GroupId: "1", MemberIds: []string{"2"}, IsAdmin: true})
	require.NoError(t, err)
	var receipt objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ?", receiptTypeGroup, request.ID, "2").First(&receipt).Error)
	require.Zero(t, receipt.IsRead)
	require.Equal(t, 1, receipt.IsActionable)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))

	_, err = NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{UserId: "1", ActorUid: "1", GroupId: "1", MemberIds: []string{"2"}, IsAdmin: false})
	require.NoError(t, err)
	require.NoError(t, svcCtx.DB.First(&receipt, receipt.ID).Error)
	require.Zero(t, receipt.IsActionable)
	received, err := NewGetGroupPutListByUidLogic(context.Background(), svcCtx).GetGroupPutListByUid(&social.GetGroupPutListByUidReq{Ids: []string{"2"}, ActorUid: "2", Class: "2"})
	require.NoError(t, err)
	require.Equal(t, int64(1), received.Total)
	require.False(t, received.List[0].Actionable)
}

func TestGroupAdminReceiptFailureRollsBackRole(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger injects receipt failure")
	}
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 0)
	now := time.Now()
	request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&request).Error)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_admin_receipt BEFORE INSERT ON social_request_receipts BEGIN SELECT RAISE(ABORT, 'receipt failure'); END`).Error)

	_, err := NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{UserId: "1", ActorUid: "1", GroupId: "1", MemberIds: []string{"2"}, IsAdmin: true})
	require.Equal(t, codes.Internal, status.Code(err))
	var member objects.GroupMember
	require.NoError(t, svcCtx.DB.Where("group_id = ? AND user_id = ?", 1, 2).First(&member).Error)
	require.Zero(t, member.RoleLevel)
	require.Zero(t, countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))
}

func TestGroupReadAndCountValidateActor(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	for _, actor := range []string{"", "0", "invalid"} {
		_, err := NewMarkGroupReqReadLogic(context.Background(), svcCtx).MarkGroupReqRead(&social.MarkGroupReqReadReq{UserId: actor, ActorUid: actor, RequestIds: []uint64{1}})
		require.NotEqual(t, codes.OK, status.Code(err))
		_, err = NewGroupInvitationReadLogic(context.Background(), svcCtx).GroupInvitationRead(&social.GroupInvitationReadReq{ActorUid: actor, InvitationIds: []uint64{1}})
		require.NotEqual(t, codes.OK, status.Code(err))
		_, err = NewGroupRequestMessageCountLogic(context.Background(), svcCtx).GroupRequestMessageCount(&social.GroupRequestMessageCountReq{UserId: actor, ActorUid: actor})
		require.NotEqual(t, codes.OK, status.Code(err))
	}
	_, err := NewMarkGroupReqReadLogic(context.Background(), svcCtx).MarkGroupReqRead(&social.MarkGroupReqReadReq{UserId: "1", ActorUid: "1"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGroupNotificationOutboxFailureRollsBackApplication(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger injects failure")
	}
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_group_notification BEFORE INSERT ON social_notification_outbox BEGIN SELECT RAISE(ABORT, 'notification failure'); END`).Error)
	_, err := NewGroupPutinLogic(context.Background(), svcCtx).GroupPutin(groupPutInReq("3", "3", "1"))
	require.Equal(t, codes.Internal, status.Code(err))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupRequest{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))
}

func TestGroupInvitationReadRequiresDisplayedIDs(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	_, err := NewGroupInvitationReadLogic(context.Background(), svcCtx).GroupInvitationRead(&social.GroupInvitationReadReq{ActorUid: "3"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGroupInvitationCreateAuthorizationAndIndependence(t *testing.T) {
	tests := []struct {
		name    string
		inviter bool
		invitee string
		code    codes.Code
	}{
		{name: "member may invite", inviter: true, invitee: "3"},
		{name: "non member forbidden", invitee: "3", code: codes.PermissionDenied},
		{name: "self forbidden", inviter: true, invitee: "2", code: codes.InvalidArgument},
		{name: "disabled invitee", inviter: true, invitee: "20", code: codes.NotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, recorder := newGroupTestContext(t)
			require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
			if tt.inviter {
				seedGroupMember(t, svcCtx.DB, 1, 2, 0)
			}
			resp, err := NewGroupInvitationCreateLogic(context.Background(), svcCtx).GroupInvitationCreate(&social.GroupInvitationCreateReq{
				ActorUid: "2", GroupId: "1", InviteeUid: tt.invitee, Message: "join us",
			})
			require.Equal(t, tt.code, status.Code(err))
			if tt.code != codes.OK {
				require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupInvitation{}))
				return
			}
			require.Equal(t, int32(0), resp.Invitation.Status)
			require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupInvitation{}))
			require.Zero(t, recorder.count())
		})
	}

	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 0)
	for _, inviter := range []string{"1", "2"} {
		_, err := NewGroupInvitationCreateLogic(context.Background(), svcCtx).GroupInvitationCreate(&social.GroupInvitationCreateReq{
			ActorUid: inviter, GroupId: "1", InviteeUid: "3",
		})
		require.NoError(t, err)
	}
	require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupInvitation{}))
}

func TestGroupInvitationDisabledUsers(t *testing.T) {
	t.Run("disabled inviter cannot create", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 20, 0)
		_, err := NewGroupInvitationCreateLogic(context.Background(), svcCtx).GroupInvitationCreate(&social.GroupInvitationCreateReq{
			ActorUid: "20", GroupId: "1", InviteeUid: "3",
		})
		require.Equal(t, codes.NotFound, status.Code(err))
		require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupInvitation{}))
	})

	t.Run("legacy batch is atomic for disabled users", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 2, 0)
		_, err := NewGroupInviteLogic(context.Background(), svcCtx).GroupInvite(&social.GroupInviteReq{
			ActorUid: "2", UserId: "2", GroupId: "1", FriendIds: []string{"3", "20"},
		})
		require.Equal(t, codes.NotFound, status.Code(err))
		require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupInvitation{}))
	})

	t.Run("disabled inviter cannot use legacy batch", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 20, 0)
		_, err := NewGroupInviteLogic(context.Background(), svcCtx).GroupInvite(&social.GroupInviteReq{
			ActorUid: "20", UserId: "20", GroupId: "1", FriendIds: []string{"3"},
		})
		require.Equal(t, codes.NotFound, status.Code(err))
		require.Zero(t, countRows(t, svcCtx.DB, &objects.GroupInvitation{}))
	})

	t.Run("disabled invitee cannot handle", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		invitation := seedInvitation(t, svcCtx.DB, 1, 1, 20, 2)
		_, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{
			Id: invitation.ID, ActorUid: "20", Result: 1,
		})
		require.Equal(t, codes.NotFound, status.Code(err))
		var latest objects.GroupInvitation
		require.NoError(t, svcCtx.DB.First(&latest, invitation.ID).Error)
		require.Equal(t, groupInvitationPending, latest.Status)
	})
}

func TestGroupInvitationHandleRolesAndAuthorization(t *testing.T) {
	tests := []struct {
		name           string
		roleAtCreate   int
		roleAtHandle   *int
		actor          string
		result         int32
		wantState      string
		wantStatus     int32
		wantMember     int64
		wantRequest    int64
		wantInvitation int32
		code           codes.Code
	}{
		{name: "only invitee", roleAtCreate: 2, actor: "4", result: 1, code: codes.PermissionDenied},
		{name: "reject only current", roleAtCreate: 0, actor: "3", result: 2, wantState: "rejected", wantStatus: 2, wantMember: 1, wantInvitation: 2},
		{name: "admin joins", roleAtCreate: 1, actor: "3", result: 1, wantState: "joined", wantStatus: 1, wantMember: 2, wantInvitation: 1},
		{name: "member approval", roleAtCreate: 0, actor: "3", result: 1, wantState: "pending_approval", wantStatus: 1, wantMember: 1, wantRequest: 1, wantInvitation: 1},
		{name: "role promoted", roleAtCreate: 0, roleAtHandle: intPtr(1), actor: "3", result: 1, wantState: "joined", wantStatus: 1, wantMember: 2, wantInvitation: 1},
		{name: "role demoted", roleAtCreate: 1, roleAtHandle: intPtr(0), actor: "3", result: 1, wantState: "pending_approval", wantStatus: 1, wantMember: 1, wantRequest: 1, wantInvitation: 1},
		{name: "inviter left", roleAtCreate: 1, roleAtHandle: intPtr(-1), actor: "3", result: 1, wantState: "invalidated", wantStatus: 3, wantInvitation: 3},
		{name: "invalid result", roleAtCreate: 1, actor: "3", result: 3, code: codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, _ := newGroupTestContext(t)
			require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
			seedGroupMember(t, svcCtx.DB, 1, 2, tt.roleAtCreate)
			invitation := seedInvitation(t, svcCtx.DB, 1, 2, 3, tt.roleAtCreate)
			if tt.roleAtHandle != nil {
				if *tt.roleAtHandle < 0 {
					require.NoError(t, svcCtx.DB.Where("group_id = ? AND user_id = ?", 1, 2).Delete(&objects.GroupMember{}).Error)
				} else {
					require.NoError(t, svcCtx.DB.Model(&objects.GroupMember{}).Where("group_id = ? AND user_id = ?", 1, 2).Update("role_level", *tt.roleAtHandle).Error)
				}
			}
			resp, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{
				Id: invitation.ID, ActorUid: tt.actor, Result: tt.result, HandleMsg: "no",
			})
			require.Equal(t, tt.code, status.Code(err))
			if tt.code != codes.OK {
				return
			}
			require.Equal(t, tt.wantState, resp.JoinState)
			require.Equal(t, tt.wantStatus, resp.Status)
			require.Equal(t, tt.wantMember, countRows(t, svcCtx.DB, &objects.GroupMember{}))
			require.Equal(t, tt.wantRequest, countRows(t, svcCtx.DB, &objects.GroupRequest{}))
			var latest objects.GroupInvitation
			require.NoError(t, svcCtx.DB.First(&latest, invitation.ID).Error)
			require.Equal(t, int(tt.wantInvitation), latest.Status)
		})
	}
}

func TestGroupInvitationAcceptCreatesIndependentApprovalAndInvalidatesOthers(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 0)
	now := time.Now()
	active := "group:direct:1:3"
	direct := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(1), HandleResult: intPtr(0), ActiveKey: &active, SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&direct).Error)
	first := seedInvitation(t, svcCtx.DB, 1, 2, 3, 0)
	second := seedInvitation(t, svcCtx.DB, 1, 1, 3, 2)
	resp, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{Id: first.ID, ActorUid: "3", Result: 1})
	require.NoError(t, err)
	require.Equal(t, "pending_approval", resp.JoinState)
	require.NotZero(t, resp.GroupRequestId)
	require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupRequest{}))
	var latest objects.GroupInvitation
	require.NoError(t, svcCtx.DB.First(&latest, second.ID).Error)
	require.Equal(t, groupInvitationInvalidated, latest.Status)

	retry, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{Id: first.ID, ActorUid: "3", Result: 1})
	require.NoError(t, err)
	require.True(t, retry.Idempotent)
	require.Equal(t, "pending_approval", retry.JoinState)
	require.Equal(t, resp.GroupRequestId, retry.GroupRequestId)
	_, err = NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{Id: first.ID, ActorUid: "3", Result: 2})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestGroupInvitationRetryReflectsApprovalTerminalState(t *testing.T) {
	for _, tc := range []struct {
		name      string
		result    int32
		joinState string
	}{
		{name: "approved", result: groupRequestAccepted, joinState: "joined"},
		{name: "rejected", result: groupRequestRejected, joinState: "approval_rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svcCtx, _ := newGroupTestContext(t)
			require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
			seedGroupMember(t, svcCtx.DB, 1, 1, 2)
			seedGroupMember(t, svcCtx.DB, 1, 2, 0)
			invitation := seedInvitation(t, svcCtx.DB, 1, 2, 3, 0)
			accepted, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{
				Id: invitation.ID, ActorUid: "3", Result: 1,
			})
			require.NoError(t, err)
			require.Equal(t, "pending_approval", accepted.JoinState)

			_, err = NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
				RequestId: accepted.GroupRequestId, ActorUid: "1", HandleResult: tc.result,
			})
			require.NoError(t, err)
			retry, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{
				Id: invitation.ID, ActorUid: "3", Result: 1,
			})
			require.NoError(t, err)
			require.True(t, retry.Idempotent)
			require.Equal(t, tc.joinState, retry.JoinState)
			require.Equal(t, accepted.GroupRequestId, retry.GroupRequestId)
			require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupRequest{}))
		})
	}
}

func TestGroupInvitationConcurrentAcceptSingleWinner(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 1)
	first := seedInvitation(t, svcCtx.DB, 1, 1, 3, 2)
	second := seedInvitation(t, svcCtx.DB, 1, 2, 3, 1)
	var wg sync.WaitGroup
	for _, id := range []uint64{first.ID, second.ID} {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			_, _ = NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{Id: id, ActorUid: "3", Result: 1})
		}(id)
	}
	wg.Wait()
	var invitations []objects.GroupInvitation
	require.NoError(t, svcCtx.DB.Order("id").Find(&invitations).Error)
	statuses := []int{invitations[0].Status, invitations[1].Status}
	require.Contains(t, statuses, groupInvitationAccepted)
	require.Contains(t, statuses, groupInvitationInvalidated)
	require.Equal(t, int64(3), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
}

func TestGroupInvitationExpiryAndListScope(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	expired := seedInvitation(t, svcCtx.DB, 1, 1, 3, 2)
	require.NoError(t, svcCtx.DB.Model(&objects.GroupInvitation{}).Where("id = ?", expired.ID).Update("expires_at", time.Now().Add(-time.Second)).Error)
	other := seedInvitation(t, svcCtx.DB, 1, 1, 4, 2)

	resp, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{
		Id: expired.ID, ActorUid: "3", Result: 1,
	})
	require.NoError(t, err)
	require.Equal(t, int32(groupInvitationExpired), resp.Status)
	require.Equal(t, "expired", resp.JoinState)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))

	list, err := NewGroupInvitationListLogic(context.Background(), svcCtx).GroupInvitationList(&social.GroupInvitationListReq{
		ActorUid: "3", Status: -1, Page: 1, Size: 20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), list.Total)
	require.Len(t, list.List, 1)
	require.Equal(t, expired.ID, list.List[0].Id)
	require.NotEqual(t, other.ID, list.List[0].Id)
}

func TestGroupRequestHandleCASClosureAndRollback(t *testing.T) {
	svcCtx, recorder := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 0)
	now := time.Now().Add(-time.Minute)
	active := "group:direct:1:3"
	first := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(1), HandleResult: intPtr(0), ActiveKey: &active, SourceType: 1}
	second := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(2), HandleResult: intPtr(0), SourceType: 2}
	require.NoError(t, svcCtx.DB.Create(&first).Error)
	require.NoError(t, svcCtx.DB.Create(&second).Error)
	invitation := seedInvitation(t, svcCtx.DB, 1, 2, 3, 0)

	for _, tc := range []struct {
		name   string
		actor  string
		result int32
		code   codes.Code
	}{
		{name: "member forbidden", actor: "2", result: 1, code: codes.PermissionDenied},
		{name: "invalid result", actor: "1", result: 3, code: codes.InvalidArgument},
		{name: "owner accepts", actor: "1", result: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
				GroupReqId: int32(first.ID), ActorUid: tc.actor, HandleUid: tc.actor, HandleResult: tc.result,
			})
			require.Equal(t, tc.code, status.Code(err))
			if tc.code == codes.OK {
				require.False(t, resp.Idempotent)
			}
		})
	}
	var requests []objects.GroupRequest
	require.NoError(t, svcCtx.DB.Find(&requests).Error)
	for _, request := range requests {
		require.Equal(t, groupRequestAccepted, *request.HandleResult)
		require.Nil(t, request.ActiveKey)
	}
	var latestInvitation objects.GroupInvitation
	require.NoError(t, svcCtx.DB.First(&latestInvitation, invitation.ID).Error)
	require.Equal(t, groupInvitationInvalidated, latestInvitation.Status)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	require.Zero(t, recorder.count())

	retry, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{GroupReqId: int32(first.ID), ActorUid: "1", HandleUid: "1", HandleResult: 1})
	require.NoError(t, err)
	require.True(t, retry.Idempotent)
	_, err = NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{GroupReqId: int32(first.ID), ActorUid: "1", HandleUid: "1", HandleResult: 2})
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestGroupRequestIDCompatibility(t *testing.T) {
	tests := []struct {
		name   string
		legacy int32
		modern uint64
		code   codes.Code
	}{
		{name: "legacy only", legacy: 1},
		{name: "modern only", modern: 1},
		{name: "matching fields", legacy: 1, modern: 1},
		{name: "conflicting fields", legacy: 1, modern: 2, code: codes.InvalidArgument},
		{name: "missing fields", code: codes.InvalidArgument},
		{name: "negative legacy", legacy: -1, code: codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, _ := newGroupTestContext(t)
			require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
			seedGroupMember(t, svcCtx.DB, 1, 1, 2)
			now := time.Now()
			request := objects.GroupRequest{ID: 1, ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
			require.NoError(t, svcCtx.DB.Create(&request).Error)
			_, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
				GroupReqId: tt.legacy, RequestId: tt.modern, ActorUid: "1", HandleResult: 2,
			})
			require.Equal(t, tt.code, status.Code(err))
		})
	}

	t.Run("modern id above int32", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		now := time.Now()
		request := objects.GroupRequest{ID: uint64(1<<31) + 1, ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
		require.NoError(t, svcCtx.DB.Create(&request).Error)
		resp, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
			RequestId: request.ID, ActorUid: "1", HandleResult: 2,
		})
		require.NoError(t, err)
		require.Equal(t, request.ID, resp.RequestId)
	})
}

func TestGroupKickSkippedDoesNotWriteOutbox(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 2)
	_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{
		UserId: "1", GroupId: "1", MemberIds: []string{"2", "3"},
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
}

func TestGroupKickDeduplicatesMembers(t *testing.T) {
	svcCtx, recorder := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 0)
	_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{
		UserId: "1", GroupId: "1", MemberIds: []string{"2", "2"},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	require.Equal(t, 1, recorder.count())
}

func TestGroupKickBatchIsAtomicAndAuthorized(t *testing.T) {
	t.Run("empty batch still requires authorization", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 2, 0)
		_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{
			UserId: "2", GroupId: "1",
		})
		require.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("invalid later member changes nothing", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		seedGroupMember(t, svcCtx.DB, 1, 2, 0)
		_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{
			UserId: "1", GroupId: "1", MemberIds: []string{"2", "invalid"},
		})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
		require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupMember{}))
		require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	})

	t.Run("outbox failure rolls back all members", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		if svcCtx.DB.Dialector.Name() != "sqlite" {
			t.Skip("SQLite trigger is used to inject the outbox failure")
		}
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		seedGroupMember(t, svcCtx.DB, 1, 2, 0)
		seedGroupMember(t, svcCtx.DB, 1, 3, 0)
		require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_kick_outbox BEFORE INSERT ON relation_outbox BEGIN SELECT RAISE(ABORT, 'outbox failure'); END`).Error)
		_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{
			UserId: "1", GroupId: "1", MemberIds: []string{"2", "3"},
		})
		require.Equal(t, codes.Internal, status.Code(err))
		require.Equal(t, int64(3), countRows(t, svcCtx.DB, &objects.GroupMember{}))
		require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	})

	t.Run("successful batch writes one outbox per member", func(t *testing.T) {
		svcCtx, recorder := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		seedGroupMember(t, svcCtx.DB, 1, 2, 0)
		seedGroupMember(t, svcCtx.DB, 1, 3, 0)
		_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{
			UserId: "1", GroupId: "1", MemberIds: []string{"2", "3"},
		})
		require.NoError(t, err)
		require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
		require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
		require.Equal(t, 2, recorder.count())
	})
}

func TestGroupMembershipRemovalClosesAdminReceipts(t *testing.T) {
	for _, tt := range []struct {
		name   string
		remove func(*svc.ServiceContext) error
	}{
		{name: "kick", remove: func(svcCtx *svc.ServiceContext) error {
			_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{UserId: "1", GroupId: "1", MemberIds: []string{"2"}})
			return err
		}},
		{name: "quit", remove: func(svcCtx *svc.ServiceContext) error {
			_, err := NewGroupQuitLogic(context.Background(), svcCtx).GroupQuit(&social.GroupQuitReq{UserId: "2", GroupId: "1"})
			return err
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, _ := newGroupTestContext(t)
			require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
			seedGroupMember(t, svcCtx.DB, 1, 1, 2)
			seedGroupMember(t, svcCtx.DB, 1, 2, 1)
			now := time.Now()
			request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
			require.NoError(t, svcCtx.DB.Create(&request).Error)
			require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroup, request.ID, "2", receiptKindApply, 1, receiptPending, now, false, nil))

			require.NoError(t, tt.remove(svcCtx))
			var receipt objects.SocialRequestReceipt
			require.NoError(t, svcCtx.DB.Where("request_id = ? AND receiver_id = ?", request.ID, "2").First(&receipt).Error)
			require.Zero(t, receipt.IsActionable)
			var members int64
			require.NoError(t, svcCtx.DB.Model(&objects.GroupMember{}).Where("group_id = ? AND user_id = ?", 1, 2).Count(&members).Error)
			require.Zero(t, members)
		})
	}
}

func TestGroupMembershipRemovalReceiptFailureRollsBack(t *testing.T) {
	for _, tt := range []struct {
		name   string
		remove func(*svc.ServiceContext) error
	}{
		{name: "kick", remove: func(svcCtx *svc.ServiceContext) error {
			_, err := NewGroupKickLogic(context.Background(), svcCtx).GroupKick(&social.GroupKickReq{UserId: "1", GroupId: "1", MemberIds: []string{"2"}})
			return err
		}},
		{name: "quit", remove: func(svcCtx *svc.ServiceContext) error {
			_, err := NewGroupQuitLogic(context.Background(), svcCtx).GroupQuit(&social.GroupQuitReq{UserId: "2", GroupId: "1"})
			return err
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, _ := newGroupTestContext(t)
			if svcCtx.DB.Dialector.Name() != "sqlite" {
				t.Skip("SQLite trigger injects receipt failure")
			}
			require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
			seedGroupMember(t, svcCtx.DB, 1, 1, 2)
			seedGroupMember(t, svcCtx.DB, 1, 2, 1)
			now := time.Now()
			request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
			require.NoError(t, svcCtx.DB.Create(&request).Error)
			require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroup, request.ID, "2", receiptKindApply, 1, receiptPending, now, false, nil))
			require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_remove_admin_receipt BEFORE UPDATE ON social_request_receipts BEGIN SELECT RAISE(ABORT, 'receipt failure'); END`).Error)

			require.Equal(t, codes.Internal, status.Code(tt.remove(svcCtx)))
			var members int64
			require.NoError(t, svcCtx.DB.Model(&objects.GroupMember{}).Where("group_id = ? AND user_id = ?", 1, 2).Count(&members).Error)
			require.Equal(t, int64(1), members)
			var receipt objects.SocialRequestReceipt
			require.NoError(t, svcCtx.DB.Where("request_id = ? AND receiver_id = ?", request.ID, "2").First(&receipt).Error)
			require.Equal(t, 1, receipt.IsActionable)
		})
	}
}

func TestGroupRoleRevocationSerializesWithRequestApproval(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 1)
	now := time.Now()
	request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&request).Error)

	start := make(chan struct{})
	errs := make(chan error, 2)
	go func() {
		<-start
		_, err := NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{
			UserId: "1", ActorUid: "1", GroupId: "1", MemberIds: []string{"2"}, IsAdmin: false,
		})
		errs <- err
	}()
	go func() {
		<-start
		_, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
			RequestId: request.ID, ActorUid: "2", HandleResult: 1,
		})
		errs <- err
	}()
	close(start)
	firstErr, secondErr := <-errs, <-errs
	for _, err := range []error{firstErr, secondErr} {
		if err != nil {
			require.Equal(t, codes.PermissionDenied, status.Code(err))
		}
	}

	member, err := loadGroupMember(svcCtx.DB, 1, 2)
	require.NoError(t, err)
	require.NotNil(t, member)
	require.Zero(t, member.RoleLevel)
	var latest objects.GroupRequest
	require.NoError(t, svcCtx.DB.First(&latest, request.ID).Error)
	require.NotNil(t, latest.HandleResult)
	if *latest.HandleResult == groupRequestAccepted {
		require.Equal(t, int64(3), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	} else {
		require.Equal(t, groupRequestPending, *latest.HandleResult)
		require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	}
}

func TestGroupRoleRevocationSerializesWithInvitationConfirmation(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	seedGroupMember(t, svcCtx.DB, 1, 2, 1)
	invitation := seedInvitation(t, svcCtx.DB, 1, 2, 3, 1)

	start := make(chan struct{})
	roleErr := make(chan error, 1)
	handleResult := make(chan *social.GroupInvitationHandleResp, 1)
	handleErr := make(chan error, 1)
	go func() {
		<-start
		_, err := NewGroupSetAdminLogic(context.Background(), svcCtx).GroupSetAdmin(&social.GroupSetAdminReq{
			UserId: "1", ActorUid: "1", GroupId: "1", MemberIds: []string{"2"}, IsAdmin: false,
		})
		roleErr <- err
	}()
	go func() {
		<-start
		resp, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{
			Id: invitation.ID, ActorUid: "3", Result: 1,
		})
		handleResult <- resp
		handleErr <- err
	}()
	close(start)
	require.NoError(t, <-roleErr)
	resp, err := <-handleResult, <-handleErr
	require.NoError(t, err)
	require.Contains(t, []string{"joined", "pending_approval"}, resp.JoinState)

	inviter, err := loadGroupMember(svcCtx.DB, 1, 2)
	require.NoError(t, err)
	require.NotNil(t, inviter)
	require.Zero(t, inviter.RoleLevel)
	if resp.JoinState == "joined" {
		invitee, err := loadGroupMember(svcCtx.DB, 1, 3)
		require.NoError(t, err)
		require.NotNil(t, invitee)
	} else {
		require.NotZero(t, resp.GroupRequestId)
		var approval objects.GroupRequest
		require.NoError(t, svcCtx.DB.First(&approval, resp.GroupRequestId).Error)
		require.Equal(t, groupRequestPending, *approval.HandleResult)
	}
}

func TestGroupRequestRejectOnlyCurrentAndOutboxRollback(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	now := time.Now()
	first := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(1), HandleResult: intPtr(0), SourceType: 1}
	second := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(2), HandleResult: intPtr(0), SourceType: 2}
	require.NoError(t, svcCtx.DB.Create(&first).Error)
	require.NoError(t, svcCtx.DB.Create(&second).Error)
	_, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{GroupReqId: int32(first.ID), ActorUid: "1", HandleUid: "1", HandleResult: 2})
	require.NoError(t, err)
	assertGroupRequestState(t, svcCtx.DB, first.ID, 2)
	assertGroupRequestState(t, svcCtx.DB, second.ID, 0)

	if svcCtx.DB.Dialector.Name() != "sqlite" {
		return
	}
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_group_relation_outbox BEFORE INSERT ON relation_outbox BEGIN SELECT RAISE(ABORT, 'outbox failure'); END`).Error)
	_, err = NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{GroupReqId: int32(second.ID), ActorUid: "1", HandleUid: "1", HandleResult: 1})
	require.Equal(t, codes.Internal, status.Code(err))
	assertGroupRequestState(t, svcCtx.DB, second.ID, 0)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
}

func TestGroupRequestUnavailableBecomesInvalidated(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	group := testGroup(1, true)
	require.NoError(t, svcCtx.DB.Create(group).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	now := time.Now()
	request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, HandleResult: intPtr(0), SourceType: 1}
	require.NoError(t, svcCtx.DB.Create(&request).Error)
	require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroup, request.ID, "1", receiptKindApply, 1, 0, now, false, nil))
	require.NoError(t, svcCtx.DB.Model(&objects.Group{}).Where("id = ?", 1).Update("status", 1).Error)
	resp, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{RequestId: request.ID, ActorUid: "1", HandleResult: 1})
	require.NoError(t, err)
	require.Equal(t, int32(groupRequestInvalidated), resp.HandleResult)
	var outbox objects.SocialNotificationOutbox
	require.NoError(t, svcCtx.DB.Where("notify_type = ?", NotifyGroupInvalidated).First(&outbox).Error)
	require.Equal(t, "group:1:invalidated", outbox.BizID)
}

func TestGroupRequestConcurrentResultsAndExistingMember(t *testing.T) {
	t.Run("accept reject invariant", func(t *testing.T) {
		svcCtx, recorder := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		now := time.Now()
		request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(1), HandleResult: intPtr(0), SourceType: 1}
		require.NoError(t, svcCtx.DB.Create(&request).Error)
		var wg sync.WaitGroup
		for _, result := range []int32{1, 2} {
			wg.Add(1)
			go func(result int32) {
				defer wg.Done()
				_, _ = NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
					GroupReqId: int32(request.ID), ActorUid: "1", HandleUid: "1", HandleResult: result,
				})
			}(result)
		}
		wg.Wait()
		var latest objects.GroupRequest
		require.NoError(t, svcCtx.DB.First(&latest, request.ID).Error)
		require.NotNil(t, latest.HandleResult)
		require.Contains(t, []int{1, 2}, *latest.HandleResult)
		if *latest.HandleResult == groupRequestAccepted {
			require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupMember{}))
			require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
		} else {
			require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.GroupMember{}))
			require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
		}
		require.Zero(t, recorder.count())
	})

	t.Run("existing member creates no outbox", func(t *testing.T) {
		svcCtx, _ := newGroupTestContext(t)
		require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
		seedGroupMember(t, svcCtx.DB, 1, 1, 2)
		seedGroupMember(t, svcCtx.DB, 1, 3, 0)
		now := time.Now()
		request := objects.GroupRequest{ReqID: "3", GroupID: 1, ReqTime: &now, JoinSource: intPtr(1), HandleResult: intPtr(0), SourceType: 1}
		require.NoError(t, svcCtx.DB.Create(&request).Error)
		_, err := NewGroupPutInHandleLogic(context.Background(), svcCtx).GroupPutInHandle(&social.GroupPutInHandleReq{
			GroupReqId: int32(request.ID), ActorUid: "1", HandleUid: "1", HandleResult: 1,
		})
		require.NoError(t, err)
		assertGroupRequestState(t, svcCtx.DB, request.ID, groupRequestAccepted)
		require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.GroupMember{}))
		require.Zero(t, countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	})
}

func testGroup(id uint64, verify bool) *objects.Group {
	statusValue := 0
	verifyValue := 0
	if verify {
		verifyValue = 1
	}
	return &objects.Group{ID: id, Name: fmt.Sprintf("group-%d", id), Status: &statusValue, CreatorUID: 1, GroupType: 1, IsVerify: verifyValue}
}

func testAbnormalGroup(id uint64) *objects.Group {
	group := testGroup(id, true)
	statusValue := 1
	group.Status = &statusValue
	return group
}

func groupPutInReq(actor, legacy, groupID string) *social.GroupPutinReq {
	return &social.GroupPutinReq{ActorUid: actor, ReqId: legacy, GroupId: groupID, ReqMsg: "hello", ReqTime: time.Now().Unix(), JoinSource: 1}
}

func seedGroupMember(t *testing.T, database *gorm.DB, groupID, userID uint64, role int) {
	t.Helper()
	now := time.Now()
	require.NoError(t, database.Create(&objects.GroupMember{GroupID: groupID, UserID: userID, RoleLevel: role, JoinTime: &now}).Error)
}

func seedInvitation(t *testing.T, database *gorm.DB, groupID, inviter, invitee uint64, role int) objects.GroupInvitation {
	t.Helper()
	now := time.Now()
	invitation := objects.GroupInvitation{
		GroupID: groupID, InviterUID: inviter, InviteeUID: invitee, InviterRoleSnapshot: role,
		Status: groupInvitationPending, CreatedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	require.NoError(t, database.Create(&invitation).Error)
	return invitation
}

func seedInviteLink(t *testing.T, database *gorm.DB, token string, groupID, creator uint64, maxUses uint) {
	t.Helper()
	now := time.Now()
	require.NoError(t, database.Create(&objects.GroupInviteLink{
		GroupID: groupID, Token: token, CreatedBy: creator, MaxUses: maxUses, CreatedAt: &now,
	}).Error)
}

func assertGroupRequestState(t *testing.T, database *gorm.DB, id uint64, result int) {
	t.Helper()
	var request objects.GroupRequest
	require.NoError(t, database.First(&request, id).Error)
	require.NotNil(t, request.HandleResult)
	require.Equal(t, result, *request.HandleResult)
}
