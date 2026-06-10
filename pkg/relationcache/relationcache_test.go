package relationcache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// newTestCache 连接本地真实 Redis 的 DB 15（测试专用库），连不上则 skip。
// 不 mock：遵守 test-files.md「不要 mock 数据库」。每个用例自带唯一 id，测试后 FlushDB 清理。
func newTestCache(t *testing.T) (*Cache, func()) {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   15,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("local redis unavailable, skip: %v", err)
	}
	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flush test db: %v", err)
	}
	cleanup := func() {
		rdb.FlushDB(context.Background())
		rdb.Close()
	}
	return New(rdb), cleanup
}

func Test_IsGroupMember_Verdict(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()

	// 已加载：g1 成员 {u1, u2}；g2 已加载但为空群（仅哨兵）
	if err := c.LoadGroupMembers(ctx, "g1", []string{"u1", "u2"}, 1); err != nil {
		t.Fatalf("load g1: %v", err)
	}
	if err := c.LoadGroupMembers(ctx, "g2", []string{}, 1); err != nil {
		t.Fatalf("load g2: %v", err)
	}

	tests := []struct {
		name string
		gid  string
		uid  string
		want Verdict
	}{
		{"loaded member -> allowed", "g1", "u1", VerdictAllowed},
		{"loaded non-member -> denied", "g1", "u9", VerdictDenied},
		{"loaded empty group, sentinel not leaked -> denied", "g2", "u1", VerdictDenied},
		{"loaded empty group, sentinel itself not a member", "g2", loadedSentinel, VerdictDenied},
		{"not loaded group -> unknown (caller fail-open)", "gX", "u1", VerdictUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.IsGroupMember(ctx, tt.gid, tt.uid); got != tt.want {
				t.Errorf("IsGroupMember(%s,%s) = %v, want %v", tt.gid, tt.uid, got, tt.want)
			}
		})
	}
}
