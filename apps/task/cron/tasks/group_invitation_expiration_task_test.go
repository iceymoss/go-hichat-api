package tasks

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"

	"github.com/stretchr/testify/require"
	zeroredis "github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
)

type fakeInvitationExpirer struct {
	mu        sync.Mutex
	responses []*social.ExpireGroupInvitationsResp
	err       error
	calls     int
	batches   []int32
	afterCall func(int)
	returnNil bool
	block     <-chan struct{}
	entered   chan<- struct{}
}

func (f *fakeInvitationExpirer) ExpireGroupInvitations(ctx context.Context, in *social.ExpireGroupInvitationsReq, _ ...grpc.CallOption) (*social.ExpireGroupInvitationsResp, error) {
	f.mu.Lock()
	f.calls++
	call := f.calls
	f.batches = append(f.batches, in.BatchSize)
	var response *social.ExpireGroupInvitationsResp
	if call <= len(f.responses) {
		response = f.responses[call-1]
	}
	afterCall := f.afterCall
	err := f.err
	returnNil := f.returnNil
	block := f.block
	entered := f.entered
	f.mu.Unlock()
	if entered != nil {
		entered <- struct{}{}
	}
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if afterCall != nil {
		afterCall(call)
	}
	if err != nil {
		return nil, err
	}
	if returnNil {
		return nil, nil
	}
	if response == nil {
		response = &social.ExpireGroupInvitationsResp{}
	}
	return response, nil
}

type errorLocker struct{ err error }

func (l errorLocker) SetnxExCtx(context.Context, string, string, int) (bool, error) {
	return false, l.err
}

type unlockErrorLocker struct {
	svc.RedisLocker
	err error
}

func (l unlockErrorLocker) EvalCtx(context.Context, string, []string, ...any) (any, error) {
	return nil, l.err
}
func (l errorLocker) EvalCtx(context.Context, string, []string, ...any) (any, error) {
	return nil, nil
}

func TestGroupInvitationExpirationTaskBatchesAndUnlocks(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
	require.NoError(t, err)
	fake := &fakeInvitationExpirer{responses: []*social.ExpireGroupInvitationsResp{
		{Expired: 3, HasMore: true},
		{Expired: 1, HasMore: false},
	}}
	task := newExpirationTask(fake, client)
	require.NoError(t, task.Execute(context.Background()))
	require.Equal(t, 2, fake.calls)
	require.Equal(t, []int32{3, 3}, fake.batches)
	require.False(t, server.Exists(groupInvitationExpirationLockKey))
}

func TestGroupInvitationExpirationTaskLockExclusion(t *testing.T) {
	server := miniredis.RunT(t)
	server.Set(groupInvitationExpirationLockKey, "other")
	client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
	require.NoError(t, err)
	fake := &fakeInvitationExpirer{}
	require.NoError(t, newExpirationTask(fake, client).Execute(context.Background()))
	require.Zero(t, fake.calls)
	value, err := server.Get(groupInvitationExpirationLockKey)
	require.NoError(t, err)
	require.Equal(t, "other", value)
}

func TestGroupInvitationExpirationTaskCancellationAndErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*fakeInvitationExpirer, context.CancelFunc)
		wantErr string
	}{
		{name: "cancellation", setup: func(fake *fakeInvitationExpirer, cancel context.CancelFunc) {
			fake.responses = []*social.ExpireGroupInvitationsResp{{Expired: 1, HasMore: true}}
			fake.afterCall = func(int) { cancel() }
		}, wantErr: context.Canceled.Error()},
		{name: "rpc error", setup: func(fake *fakeInvitationExpirer, _ context.CancelFunc) {
			fake.err = errors.New("rpc unavailable")
		}, wantErr: "rpc unavailable"},
		{name: "nil response", setup: func(fake *fakeInvitationExpirer, _ context.CancelFunc) {
			fake.returnNil = true
		}, wantErr: "nil response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := miniredis.RunT(t)
			client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
			require.NoError(t, err)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fake := &fakeInvitationExpirer{}
			tt.setup(fake, cancel)
			err = newExpirationTask(fake, client).Execute(ctx)
			require.ErrorContains(t, err, tt.wantErr)
			require.False(t, server.Exists(groupInvitationExpirationLockKey))
		})
	}
}

func TestGroupInvitationExpirationTaskContinuesAfterConcurrentCASMisses(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
	require.NoError(t, err)
	fake := &fakeInvitationExpirer{responses: []*social.ExpireGroupInvitationsResp{
		{Expired: 0, HasMore: true},
		{Expired: 0, HasMore: false},
	}}
	require.NoError(t, newExpirationTask(fake, client).Execute(context.Background()))
	require.Equal(t, 2, fake.calls)
}

func TestGroupInvitationExpirationTaskBlockedRPCCancellation(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
	require.NoError(t, err)
	entered := make(chan struct{}, 1)
	fake := &fakeInvitationExpirer{block: make(chan struct{}), entered: entered}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- newExpirationTask(fake, client).Execute(ctx) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("RPC was not entered")
	}
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("blocked RPC did not observe cancellation")
	}
}

func TestGroupInvitationExpirationTaskConcurrentExecuteSingleRPC(t *testing.T) {
	server := miniredis.RunT(t)
	client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
	require.NoError(t, err)
	release := make(chan struct{})
	entered := make(chan struct{}, 2)
	fake := &fakeInvitationExpirer{block: release, entered: entered}
	task := newExpirationTask(fake, client)
	errs := make(chan error, 2)
	go func() { errs <- task.Execute(context.Background()) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first RPC was not entered")
	}
	go func() { errs <- task.Execute(context.Background()) }()
	select {
	case <-entered:
		t.Fatal("second execution entered RPC while lock was held")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	require.Equal(t, 1, fake.calls)
}

func TestGroupInvitationExpirationTaskCompareDeleteAndLockError(t *testing.T) {
	t.Run("does not delete replacement token", func(t *testing.T) {
		server := miniredis.RunT(t)
		client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
		require.NoError(t, err)
		fake := &fakeInvitationExpirer{afterCall: func(int) { server.Set(groupInvitationExpirationLockKey, "replacement") }}
		require.NoError(t, newExpirationTask(fake, client).Execute(context.Background()))
		value, err := server.Get(groupInvitationExpirationLockKey)
		require.NoError(t, err)
		require.Equal(t, "replacement", value)
	})
	t.Run("acquisition error propagates", func(t *testing.T) {
		fake := &fakeInvitationExpirer{}
		task := newExpirationTask(fake, errorLocker{err: errors.New("redis unavailable")})
		require.ErrorContains(t, task.Execute(context.Background()), "redis unavailable")
		require.Zero(t, fake.calls)
	})
	t.Run("work error takes precedence over unlock error", func(t *testing.T) {
		server := miniredis.RunT(t)
		client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
		require.NoError(t, err)
		fake := &fakeInvitationExpirer{err: errors.New("rpc failed")}
		err = newExpirationTask(fake, unlockErrorLocker{RedisLocker: client, err: errors.New("unlock failed")}).Execute(context.Background())
		require.ErrorContains(t, err, "rpc failed")
		require.NotContains(t, err.Error(), "unlock failed")
	})
	t.Run("unlock error returned after successful work", func(t *testing.T) {
		server := miniredis.RunT(t)
		client, err := zeroredis.NewRedis(zeroredis.RedisConf{Host: server.Addr(), Type: "node"})
		require.NoError(t, err)
		err = newExpirationTask(&fakeInvitationExpirer{}, unlockErrorLocker{RedisLocker: client, err: errors.New("unlock failed")}).Execute(context.Background())
		require.ErrorContains(t, err, "unlock failed")
	})
}

func newExpirationTask(socialClient svc.SocialInvitationExpirer, locker svc.RedisLocker) *GroupInvitationExpirationTask {
	var c config.Config
	c.Cron.InvitationExpirationSpec = "0 * * * * *"
	c.Cron.BatchSize = 3
	c.Cron.TaskTimeout = int((30 * time.Second) / time.Second)
	c.Cron.LockTTLSeconds = 60
	return NewGroupInvitationExpirationTask(&svc.ServiceContext{Config: c, Social: socialClient, Redis: locker})
}
