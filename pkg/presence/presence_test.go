package presence

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestStoreOwnerSafeLifecycle(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	store := New(client)
	ctx := context.Background()
	require.NoError(t, store.Claim(ctx, "7", "node-a", "token-a", time.Minute))
	value, err := client.Get(ctx, Key("7")).Result()
	require.NoError(t, err)
	require.Equal(t, "node-a", value)
	require.NoError(t, store.Claim(ctx, "7", "node-b", "token-b", time.Minute))
	owned, err := store.Refresh(ctx, "7", "token-a", time.Minute)
	require.NoError(t, err)
	require.False(t, owned)
	deleted, err := store.DeleteIfOwner(ctx, "7", "token-a")
	require.NoError(t, err)
	require.False(t, deleted)
	online, err := store.BatchOnline(ctx, []string{"7", "8"})
	require.NoError(t, err)
	require.Equal(t, map[string]bool{"7": true, "8": false}, online)
	deleted, err = store.DeleteIfOwner(ctx, "7", "token-b")
	require.NoError(t, err)
	require.True(t, deleted)
}
