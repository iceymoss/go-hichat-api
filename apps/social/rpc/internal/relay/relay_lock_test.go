package relay

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestRenewRelationLockIsOwnerSafe(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	ctx, cancel := context.WithCancel(context.Background())
	lost := make(chan struct{})
	originalInterval := relationLockRenewInterval
	relationLockRenewInterval = 10 * time.Millisecond
	t.Cleanup(func() { relationLockRenewInterval = originalInterval })
	require.NoError(t, client.Set(ctx, lockKey, "owner-a", lockTTL).Err())
	go renewRelationLock(ctx, client, "owner-a", lost)
	require.NoError(t, client.Set(ctx, lockKey, "owner-b", lockTTL).Err())
	select {
	case <-lost:
	case <-time.After(time.Second):
		t.Fatal("lock loss was not detected")
	}
	cancel()
	value, err := client.Get(context.Background(), lockKey).Result()
	require.NoError(t, err)
	require.Equal(t, "owner-b", value)
}
