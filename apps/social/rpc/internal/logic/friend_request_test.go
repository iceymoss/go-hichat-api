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
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type userLookupStub struct {
	users  map[string]*user.UserEntity
	errors map[string]error
}

func TestFriendRequestListPaginationAndFields(t *testing.T) {
	svcCtx, _ := newFriendTestContext(t)
	now := time.Now()
	for i, result := range []int{0, 1, 2} {
		id := uint64(math.MaxInt32) + uint64(i) + 1
		handled := now.Add(time.Duration(i) * time.Second)
		req := objects.FriendRequest{ID: id, UserID: 1, ReqUID: 2, ReqMsg: fmt.Sprintf("message-%d", i), ReqTime: now, HandleResult: intPtr(result), Status: intPtr(1), HandleMsg: "handled", HandledAt: &handled}
		require.NoError(t, svcCtx.DB.Create(&req).Error)
		require.NoError(t, createReceipt(svcCtx.DB, receiptTypeFriend, id, "2", receiptKindApply, 1, result, now, false, nil))
	}
	statusFilter := int32(0)
	filtered, err := NewFriendPutInListLogic(context.Background(), svcCtx).FriendPutInList(&social.FriendPutInListReq{UserId: "2", ActorUid: "2", Class: "1", Status: &statusFilter, Page: 1, Size: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), filtered.Total)
	require.Greater(t, filtered.List[0].RequestId, uint64(math.MaxInt32))
	require.Equal(t, "1", filtered.List[0].PeerUid)
	require.Equal(t, "handled", filtered.List[0].HandleMsg)
	all, err := NewFriendPutInListLogic(context.Background(), svcCtx).FriendPutInList(&social.FriendPutInListReq{UserId: "2", ActorUid: "2", Class: "1", Page: 2, Size: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), all.Total)
	require.Len(t, all.List, 1)
}

func (s *userLookupStub) GetUserById(_ context.Context, in *user.GetUserByIdRequest, _ ...grpc.CallOption) (*user.GetUserByIdResponse, error) {
	if err := s.errors[in.Id]; err != nil {
		return nil, err
	}
	return &user.GetUserByIdResponse{User: s.users[in.Id]}, nil
}

type notifyRecorder struct {
	mu   sync.Mutex
	msgs []*mq.CommonNotify
}

func (r *notifyRecorder) Push(msg *mq.CommonNotify) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.msgs = append(r.msgs, msg)
	return nil
}

func (r *notifyRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.msgs)
}

func newFriendTestContext(t *testing.T) (*svc.ServiceContext, *notifyRecorder) {
	t.Helper()
	db := openFriendTestDB(t)
	if db.Dialector.Name() != "sqlite" {
		requireDedicatedTestDatabase(t, db)
		resetFriendTestTables(t, db)
	}
	if db.Dialector.Name() == "sqlite" {
		require.NoError(t, db.Exec(`CREATE TABLE friend_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, req_uid INTEGER NOT NULL,
			req_msg TEXT, req_time DATETIME NOT NULL, handle_result INTEGER, handle_msg TEXT,
			handled_at DATETIME, status INTEGER, read_state INTEGER NOT NULL DEFAULT 0,
			receiver_read INTEGER NOT NULL DEFAULT 0, sender_read INTEGER NOT NULL DEFAULT 0,
			remark TEXT NOT NULL DEFAULT '', active_key TEXT UNIQUE
		)`).Error)
		require.NoError(t, db.Exec(`CREATE TABLE friends (
			id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, friend_uid INTEGER NOT NULL,
			remark TEXT, add_source INTEGER, blacklisted INTEGER NOT NULL DEFAULT 0,
			moments_permission INTEGER NOT NULL DEFAULT 0, notify_enabled INTEGER NOT NULL DEFAULT 1,
			pinned INTEGER NOT NULL DEFAULT 0, muted INTEGER NOT NULL DEFAULT 0, friend_tags TEXT,
			created_at DATETIME, UNIQUE(user_id, friend_uid)
		)`).Error)
		require.NoError(t, db.Exec(`CREATE TABLE relation_outbox (
			id INTEGER PRIMARY KEY AUTOINCREMENT, event_type TEXT NOT NULL, group_id TEXT NOT NULL DEFAULT '',
			payload TEXT NOT NULL, status INTEGER NOT NULL DEFAULT 0, created_at DATETIME, sent_at DATETIME
		)`).Error)
		require.NoError(t, db.Exec(`CREATE TABLE social_request_receipts (
			id INTEGER PRIMARY KEY AUTOINCREMENT, request_type TEXT NOT NULL, request_id INTEGER NOT NULL,
			receiver_id TEXT NOT NULL, receipt_kind TEXT NOT NULL, is_read INTEGER NOT NULL DEFAULT 0,
			is_actionable INTEGER NOT NULL DEFAULT 0, result INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL, read_at DATETIME, resolved_at DATETIME,
			UNIQUE(request_type, request_id, receiver_id, receipt_kind)
		)`).Error)
		require.NoError(t, db.AutoMigrate(&objects.SocialNotificationOutbox{}))
	} else if db.Dialector.Name() == "postgres" {
		require.NoError(t, db.AutoMigrate(&objects.FriendRequest{}))
		require.NoError(t, db.Migrator().DropIndex(&objects.FriendRequest{}, "idx_user"))
		require.NoError(t, db.AutoMigrate(&objects.Friend{}, &objects.RelationOutbox{}, &objects.SocialRequestReceipt{}, &objects.SocialNotificationOutbox{}))
	} else {
		require.NoError(t, db.AutoMigrate(&objects.FriendRequest{}, &objects.Friend{}, &objects.RelationOutbox{}, &objects.SocialRequestReceipt{}, &objects.SocialNotificationOutbox{}))
	}
	recorder := &notifyRecorder{}
	users := map[string]*user.UserEntity{
		"1": {Id: "1", Nickname: "applicant", Status: 1},
		"2": {Id: "2", Nickname: "actor", Status: 1},
		"3": {Id: "3", Nickname: "third", Status: 1},
		"4": {Id: "4", Nickname: "disabled", Status: 0},
	}
	return &svc.ServiceContext{DB: db, User: &userLookupStub{users: users}, CommonNotifyClient: recorder}, recorder
}

func resetFriendTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, table := range []string{
		"social_notification_outbox", "social_request_receipts", "relation_outbox", "friend_requests", "friends",
	} {
		require.NoError(t, db.Migrator().DropTable(table))
	}
}

func openFriendTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := "sqlite"
	dsn := "file:" + t.TempDir() + "/friend.db?_busy_timeout=10000&_journal_mode=WAL"
	dialector := gorm.Dialector(sqlite.Open(dsn))
	if value := os.Getenv("SOCIAL_FRIEND_TEST_MYSQL_DSN"); value != "" {
		name, dsn, dialector = "mysql", value, mysql.Open(value)
	} else if value := os.Getenv("SOCIAL_FRIEND_TEST_POSTGRES_DSN"); value != "" {
		name, dsn, dialector = "postgres", value, postgres.Open(value)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err, "%s DSN %s", name, dsn)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	return db
}

func TestFriendPutInValidationAndState(t *testing.T) {
	tests := []struct {
		name  string
		req   *social.FriendPutInReq
		seed  func(*testing.T, *gorm.DB)
		code  codes.Code
		check func(*testing.T, *social.FriendPutInResp, *gorm.DB, *notifyRecorder)
	}{
		{name: "missing actor", req: putInReq("", "", "2"), code: codes.Unauthenticated},
		{name: "legacy actor mismatch", req: putInReq("1", "3", "2"), code: codes.PermissionDenied},
		{name: "invalid target", req: putInReq("1", "1", "x"), code: codes.InvalidArgument},
		{name: "self", req: putInReq("1", "1", "1"), code: codes.InvalidArgument},
		{name: "missing user", req: putInReq("1", "1", "99"), code: codes.NotFound},
		{name: "user RPC not found", req: putInReq("1", "1", "98"), code: codes.NotFound},
		{name: "disabled user", req: putInReq("1", "1", "4"), code: codes.NotFound},
		{name: "already friend", req: putInReq("1", "1", "2"), seed: seedRelations(1, 2, true), check: func(t *testing.T, resp *social.FriendPutInResp, db *gorm.DB, recorder *notifyRecorder) {
			require.True(t, resp.AlreadyFriend)
			require.Equal(t, int32(1), resp.Status)
			require.Equal(t, int64(0), countRows(t, db, &objects.FriendRequest{}))
			require.Zero(t, recorder.count())
		}},
		{name: "one-way conflict", req: putInReq("1", "1", "2"), seed: seedRelations(1, 2, false), code: codes.FailedPrecondition},
		{name: "first and duplicate", req: putInReq("1", "1", "2"), check: func(t *testing.T, first *social.FriendPutInResp, db *gorm.DB, recorder *notifyRecorder) {
			require.NotZero(t, first.RequestId)
			second, err := NewFriendPutInLogic(context.Background(), &svc.ServiceContext{DB: db, User: &userLookupStub{users: normalUsers()}, CommonNotifyClient: recorder}).FriendPutIn(putInReq("1", "1", "2"))
			require.NoError(t, err)
			require.Equal(t, first.RequestId, second.RequestId)
			require.True(t, second.AlreadyPending)
			require.Equal(t, int64(1), countRows(t, db, &objects.FriendRequest{}))
			require.Zero(t, recorder.count())
			var request objects.FriendRequest
			require.NoError(t, db.First(&request, uint64(first.RequestId)).Error)
			require.Nil(t, request.HandledAt)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, recorder := newFriendTestContext(t)
			if tt.seed != nil {
				tt.seed(t, svcCtx.DB)
			}
			if tt.name == "user RPC not found" {
				svcCtx.User = &userLookupStub{users: normalUsers(), errors: map[string]error{"98": status.Error(codes.NotFound, "missing")}}
			}
			resp, err := NewFriendPutInLogic(context.Background(), svcCtx).FriendPutIn(tt.req)
			require.Equal(t, tt.code, status.Code(err))
			if tt.code != codes.OK {
				require.Equal(t, int64(0), countRows(t, svcCtx.DB, &objects.FriendRequest{}))
				return
			}
			tt.check(t, resp, svcCtx.DB, recorder)
		})
	}
}

func TestFriendPutInConcurrentActive(t *testing.T) {
	svcCtx, recorder := newFriendTestContext(t)
	const workers = 20
	var wg sync.WaitGroup
	ids := make(chan int64, workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := NewFriendPutInLogic(context.Background(), svcCtx).FriendPutIn(putInReq("1", "1", "2"))
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
	var stable int64
	for id := range ids {
		if stable == 0 {
			stable = id
		}
		require.Equal(t, stable, id)
	}
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.FriendRequest{}))
	require.Zero(t, recorder.count())
}

func TestFriendRequestReceiptsAndReadCount(t *testing.T) {
	svcCtx, _ := newFriendTestContext(t)
	created, err := NewFriendPutInLogic(context.Background(), svcCtx).FriendPutIn(&social.FriendPutInReq{
		ActorUid: "1", UserId: "1", ReqUid: "2", ReqMsg: "hello", ReqTime: time.Now().Unix(),
	})
	require.NoError(t, err)
	var apply objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ? AND receipt_kind = ?", receiptTypeFriend, created.RequestId, "2", receiptKindApply).First(&apply).Error)
	require.Zero(t, apply.IsRead)
	require.Equal(t, 1, apply.IsActionable)
	require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.SocialNotificationOutbox{}))

	count, err := NewFriendPutInMessageCountLogic(context.Background(), svcCtx).FriendPutInMessageCount(&social.FriendPutInMessageCountReq{UserId: "2", ActorUid: "2"})
	require.NoError(t, err)
	require.Equal(t, int32(1), count.Count)
	require.Equal(t, int32(1), count.Apply)

	_, err = NewFriendPutInHandleLogic(context.Background(), svcCtx).FriendPutInHandle(&social.FriendPutInHandleReq{
		FriendReqId: int32(created.RequestId), ActorUid: "2", HandleResult: 2,
	})
	require.NoError(t, err)
	require.NoError(t, svcCtx.DB.First(&apply, apply.ID).Error)
	require.Equal(t, 1, apply.IsRead)
	require.Zero(t, apply.IsActionable)
	require.Equal(t, receiptRejected, apply.Result)

	var result objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ? AND receipt_kind = ?", receiptTypeFriend, created.RequestId, "1", receiptKindResult).First(&result).Error)
	require.Zero(t, result.IsRead)
	require.Equal(t, receiptRejected, result.Result)
	require.Equal(t, int64(2), countRows(t, svcCtx.DB, &objects.SocialNotificationOutbox{}))

	read, err := NewFriendPutInReadLogic(context.Background(), svcCtx).FriendPutInRead(&social.FriendPutInReadReq{UserId: "1", ActorUid: "1", RequestIds: []uint64{uint64(created.RequestId)}})
	require.NoError(t, err)
	require.Zero(t, read.Count)
	require.NoError(t, svcCtx.DB.First(&result, result.ID).Error)
	require.Equal(t, 1, result.IsRead)
}

func TestFriendReceiptFailureRollsBackRequest(t *testing.T) {
	svcCtx, _ := newFriendTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger is used to inject receipt failure")
	}
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_friend_receipt BEFORE INSERT ON social_request_receipts BEGIN SELECT RAISE(ABORT, 'receipt failure'); END`).Error)
	_, err := NewFriendPutInLogic(context.Background(), svcCtx).FriendPutIn(&social.FriendPutInReq{
		ActorUid: "1", UserId: "1", ReqUid: "2", ReqTime: time.Now().Unix(),
	})
	require.Equal(t, codes.Internal, status.Code(err))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.FriendRequest{}))
}

func TestFriendNotificationOutboxFailureRollsBackRequest(t *testing.T) {
	svcCtx, _ := newFriendTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger injects failure")
	}
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_friend_notification BEFORE INSERT ON social_notification_outbox BEGIN SELECT RAISE(ABORT, 'notification failure'); END`).Error)
	_, err := NewFriendPutInLogic(context.Background(), svcCtx).FriendPutIn(&social.FriendPutInReq{ActorUid: "1", UserId: "1", ReqUid: "2", ReqTime: time.Now().Unix()})
	require.Equal(t, codes.Internal, status.Code(err))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.FriendRequest{}))
	require.Zero(t, countRows(t, svcCtx.DB, &objects.SocialRequestReceipt{}))
}

func TestFriendDeletePreservesSharedHistory(t *testing.T) {
	svcCtx, _ := newFriendTestContext(t)
	created, err := NewFriendPutInLogic(context.Background(), svcCtx).FriendPutIn(&social.FriendPutInReq{
		ActorUid: "1", UserId: "1", ReqUid: "2", ReqTime: time.Now().Unix(),
	})
	require.NoError(t, err)
	_, err = NewFriendPutInDeleteLogic(context.Background(), svcCtx).FriendPutInDelete(&social.FriendPutInDeleteReq{UserId: "2", ActorUid: "2", FriendReqId: int32(created.RequestId)})
	require.NoError(t, err)
	var request objects.FriendRequest
	require.NoError(t, svcCtx.DB.First(&request, uint64(created.RequestId)).Error)
	require.Equal(t, 1, *request.Status)
	require.NotNil(t, request.ActiveKey)
	var receipt objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ?", receiptTypeFriend, created.RequestId, "2").First(&receipt).Error)
	require.Equal(t, 1, receipt.IsRead)
	require.Zero(t, receipt.IsActionable)
	require.Equal(t, receiptInvalidated, receipt.Result)

	_, err = NewFriendPutInHandleLogic(context.Background(), svcCtx).FriendPutInHandle(&social.FriendPutInHandleReq{
		ActorUid: "2", UserId: "2", RequestId: uint64(created.RequestId), HandleResult: 1,
	})
	require.NoError(t, err)
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ?", receiptTypeFriend, created.RequestId, "2").First(&receipt).Error)
	require.Equal(t, receiptInvalidated, receipt.Result)
	list, err := NewFriendPutInListLogic(context.Background(), svcCtx).FriendPutInList(&social.FriendPutInListReq{UserId: "2", ActorUid: "2", Class: "1", Type: -1})
	require.NoError(t, err)
	require.Empty(t, list.List)
}

func TestFriendPutInHandleAuthorizationAndResults(t *testing.T) {
	tests := []struct {
		name   string
		actor  string
		legacy string
		result int32
		code   codes.Code
	}{
		{name: "missing actor", result: 1, code: codes.Unauthenticated},
		{name: "applicant forbidden", actor: "1", legacy: "1", result: 1, code: codes.PermissionDenied},
		{name: "third party forbidden", actor: "3", legacy: "3", result: 1, code: codes.PermissionDenied},
		{name: "legacy mismatch", actor: "2", legacy: "3", result: 1, code: codes.PermissionDenied},
		{name: "invalid result", actor: "2", legacy: "2", result: 3, code: codes.InvalidArgument},
		{name: "accept", actor: "2", legacy: "2", result: 1},
		{name: "reject", actor: "2", legacy: "2", result: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcCtx, recorder := newFriendTestContext(t)
			request := seedRequest(t, svcCtx.DB, 1, 2, "preset")
			resp, err := NewFriendPutInHandleLogic(context.Background(), svcCtx).FriendPutInHandle(&social.FriendPutInHandleReq{
				FriendReqId: int32(request.ID), ActorUid: tt.actor, UserId: tt.legacy,
				HandleResult: tt.result, HandleMsg: "handled", Remark: "reviewer", Tags: []string{"work"},
			})
			require.Equal(t, tt.code, status.Code(err))
			if tt.code != codes.OK {
				assertRequestState(t, svcCtx.DB, request.ID, 0)
				require.Zero(t, recorder.count())
				return
			}
			require.Equal(t, int64(request.ID), resp.RequestId)
			assertRequestState(t, svcCtx.DB, request.ID, int(tt.result))
			if tt.result == 1 {
				assertFriendDirections(t, svcCtx.DB)
				require.Equal(t, int64(1), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
			} else {
				require.Equal(t, int64(0), countRows(t, svcCtx.DB, &objects.Friend{}))
				require.Equal(t, int64(0), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
			}
			require.Zero(t, recorder.count())
		})
	}
}

func TestFriendPutInHandleIdempotencyAndReversePending(t *testing.T) {
	for _, result := range []int32{1, 2} {
		t.Run(fmt.Sprintf("result_%d", result), func(t *testing.T) {
			svcCtx, recorder := newFriendTestContext(t)
			request := seedRequest(t, svcCtx.DB, 1, 2, "preset")
			reverse := seedRequest(t, svcCtx.DB, 2, 1, "reverse")
			logic := NewFriendPutInHandleLogic(context.Background(), svcCtx)
			in := &social.FriendPutInHandleReq{FriendReqId: int32(request.ID), ActorUid: "2", UserId: "2", HandleResult: result}
			first, err := logic.FriendPutInHandle(in)
			require.NoError(t, err)
			require.False(t, first.Idempotent)
			second, err := logic.FriendPutInHandle(in)
			require.NoError(t, err)
			require.True(t, second.Idempotent)
			_, err = logic.FriendPutInHandle(&social.FriendPutInHandleReq{FriendReqId: int32(request.ID), ActorUid: "2", UserId: "2", HandleResult: 3 - result})
			require.Equal(t, codes.FailedPrecondition, status.Code(err))
			if result == 1 {
				assertRequestState(t, svcCtx.DB, reverse.ID, 1)
				var receipt objects.SocialRequestReceipt
				require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ? AND receiver_id = ? AND receipt_kind = ?", receiptTypeFriend, reverse.ID, "2", receiptKindResult).First(&receipt).Error)
				require.Equal(t, 1, receipt.IsRead)
			} else {
				assertRequestState(t, svcCtx.DB, reverse.ID, 0)
			}
			require.Zero(t, recorder.count())
		})
	}
}

func TestFriendPutInHandleConcurrentInvariant(t *testing.T) {
	for _, results := range [][]int32{{1, 1}, {1, 2}} {
		t.Run(fmt.Sprint(results), func(t *testing.T) {
			svcCtx, recorder := newFriendTestContext(t)
			request := seedRequest(t, svcCtx.DB, 1, 2, "")
			var wg sync.WaitGroup
			for _, result := range results {
				wg.Add(1)
				go func(result int32) {
					defer wg.Done()
					_, _ = NewFriendPutInHandleLogic(context.Background(), svcCtx).FriendPutInHandle(&social.FriendPutInHandleReq{
						FriendReqId: int32(request.ID), ActorUid: "2", UserId: "2", HandleResult: result,
					})
				}(result)
			}
			wg.Wait()
			var latest objects.FriendRequest
			require.NoError(t, svcCtx.DB.First(&latest, request.ID).Error)
			require.NotNil(t, latest.HandleResult)
			require.Contains(t, []int{1, 2}, *latest.HandleResult)
			relations := countRows(t, svcCtx.DB, &objects.Friend{})
			if *latest.HandleResult == 1 {
				require.Equal(t, int64(2), relations)
			} else {
				require.Zero(t, relations)
			}
			require.Zero(t, recorder.count())
		})
	}
}

func TestFriendPutInHandleOutboxAtomicRollback(t *testing.T) {
	svcCtx, recorder := newFriendTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("failure injection trigger is SQLite-specific")
	}
	request := seedRequest(t, svcCtx.DB, 1, 2, "")
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_relation_outbox BEFORE INSERT ON relation_outbox BEGIN SELECT RAISE(ABORT, 'outbox failure'); END`).Error)
	_, err := NewFriendPutInHandleLogic(context.Background(), svcCtx).FriendPutInHandle(&social.FriendPutInHandleReq{
		FriendReqId: int32(request.ID), ActorUid: "2", UserId: "2", HandleResult: 1,
	})
	require.Equal(t, codes.Internal, status.Code(err))
	assertRequestState(t, svcCtx.DB, request.ID, 0)
	require.Equal(t, int64(0), countRows(t, svcCtx.DB, &objects.Friend{}))
	require.Equal(t, int64(0), countRows(t, svcCtx.DB, &objects.RelationOutbox{}))
	require.Zero(t, recorder.count())
}

func putInReq(actor, legacy, target string) *social.FriendPutInReq {
	return &social.FriendPutInReq{ActorUid: actor, UserId: legacy, ReqUid: target, ReqMsg: "hello", ReqTime: time.Now().Unix(), Remark: "preset"}
}

func normalUsers() map[string]*user.UserEntity {
	return map[string]*user.UserEntity{
		"1": {Id: "1", Nickname: "applicant", Status: 1},
		"2": {Id: "2", Nickname: "actor", Status: 1},
	}
}

func seedRelations(a, b uint64, both bool) func(*testing.T, *gorm.DB) {
	return func(t *testing.T, db *gorm.DB) {
		require.NoError(t, db.Create(&objects.Friend{UserID: a, FriendUID: b}).Error)
		if both {
			require.NoError(t, db.Create(&objects.Friend{UserID: b, FriendUID: a}).Error)
		}
	}
}

func seedRequest(t *testing.T, db *gorm.DB, applicant, requested uint64, remark string) objects.FriendRequest {
	t.Helper()
	pending, visible := 0, 1
	key := "friend:" + strconv.FormatUint(applicant, 10) + ":" + strconv.FormatUint(requested, 10)
	request := objects.FriendRequest{
		UserID: applicant, ReqUID: requested, ReqMsg: "hello", ReqTime: time.Now(),
		HandleResult: &pending, Status: &visible, Remark: remark, ActiveKey: &key,
	}
	require.NoError(t, db.Create(&request).Error)
	return request
}

func countRows(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(model).Count(&count).Error)
	return count
}

func assertRequestState(t *testing.T, db *gorm.DB, id uint64, result int) {
	t.Helper()
	var request objects.FriendRequest
	require.NoError(t, db.First(&request, id).Error)
	require.NotNil(t, request.HandleResult)
	require.Equal(t, result, *request.HandleResult)
	if result == 0 {
		require.NotNil(t, request.ActiveKey)
		require.Nil(t, request.HandledAt)
	} else {
		require.Nil(t, request.ActiveKey)
		require.NotNil(t, request.HandledAt)
	}
}

func assertFriendDirections(t *testing.T, db *gorm.DB) {
	t.Helper()
	var applicantToApprover, approverToApplicant objects.Friend
	require.NoError(t, db.Where("user_id = ? AND friend_uid = ?", 1, 2).First(&applicantToApprover).Error)
	require.NoError(t, db.Where("user_id = ? AND friend_uid = ?", 2, 1).First(&approverToApplicant).Error)
	require.Equal(t, "preset", applicantToApprover.Remark)
	require.Empty(t, applicantToApprover.FriendTags)
	require.Equal(t, "reviewer", approverToApplicant.Remark)
	require.JSONEq(t, `["work"]`, approverToApplicant.FriendTags)
}

func TestFriendScopedActorValidation(t *testing.T) {
	svcCtx, _ := newFriendTestContext(t)
	for _, tc := range []struct {
		name, actor, legacy string
		code                codes.Code
	}{
		{name: "missing", legacy: "1", code: codes.Unauthenticated},
		{name: "malformed", actor: "invalid", legacy: "invalid", code: codes.InvalidArgument},
		{name: "non canonical", actor: "01", legacy: "01", code: codes.InvalidArgument},
		{name: "mismatch", actor: "1", legacy: "2", code: codes.PermissionDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewFriendPutInListLogic(context.Background(), svcCtx).FriendPutInList(&social.FriendPutInListReq{ActorUid: tc.actor, UserId: tc.legacy, Class: "0"})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewFriendPutInReadLogic(context.Background(), svcCtx).FriendPutInRead(&social.FriendPutInReadReq{ActorUid: tc.actor, UserId: tc.legacy})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewFriendPutInMessageCountLogic(context.Background(), svcCtx).FriendPutInMessageCount(&social.FriendPutInMessageCountReq{ActorUid: tc.actor, UserId: tc.legacy})
			require.Equal(t, tc.code, status.Code(err))
			_, err = NewFriendPutInDeleteLogic(context.Background(), svcCtx).FriendPutInDelete(&social.FriendPutInDeleteReq{ActorUid: tc.actor, UserId: tc.legacy, RequestId: 1})
			require.Equal(t, tc.code, status.Code(err))
		})
	}
	_, err := NewFriendPutInReadLogic(context.Background(), svcCtx).FriendPutInRead(&social.FriendPutInReadReq{ActorUid: "1", UserId: "1"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
