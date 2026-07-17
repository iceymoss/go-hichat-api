package logic

import (
	"context"
	"fmt"
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
	} else if db.Dialector.Name() == "postgres" {
		require.NoError(t, db.AutoMigrate(&objects.FriendRequest{}))
		require.NoError(t, db.Migrator().DropIndex(&objects.FriendRequest{}, "idx_user"))
		require.NoError(t, db.AutoMigrate(&objects.Friend{}, &objects.RelationOutbox{}))
	} else {
		require.NoError(t, db.AutoMigrate(&objects.FriendRequest{}, &objects.Friend{}, &objects.RelationOutbox{}))
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
			require.Equal(t, 1, recorder.count())
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
	require.Equal(t, 1, recorder.count())
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
			require.Equal(t, 1, recorder.count())
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
				var reverseRequest objects.FriendRequest
				require.NoError(t, svcCtx.DB.First(&reverseRequest, reverse.ID).Error)
				require.Equal(t, 1, reverseRequest.SenderRead)
			} else {
				assertRequestState(t, svcCtx.DB, reverse.ID, 0)
			}
			require.Equal(t, 1, recorder.count())
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
			require.Equal(t, 1, recorder.count())
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
	var applicant, actor objects.Friend
	require.NoError(t, db.Where("user_id = ? AND friend_uid = ?", 1, 2).First(&applicant).Error)
	require.NoError(t, db.Where("user_id = ? AND friend_uid = ?", 2, 1).First(&actor).Error)
	require.Equal(t, "preset", applicant.Remark)
	require.Empty(t, applicant.FriendTags)
	require.Equal(t, "reviewer", actor.Remark)
	require.JSONEq(t, `["work"]`, actor.FriendTags)
}
