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

func Test_GroupMember_VersionGate(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()

	if err := c.LoadGroupMembers(ctx, "g1", []string{"u1", "u2"}, 5); err != nil {
		t.Fatalf("load g1: %v", err)
	}

	// 新版本踢人生效
	if got, err := c.RemoveGroupMember(ctx, "g1", "u1", 6); err != nil || got != Applied {
		t.Fatalf("remove u1@6 = (%v,%v), want Applied", got, err)
	}
	if v := c.IsGroupMember(ctx, "g1", "u1"); v != VerdictDenied {
		t.Errorf("after remove u1@6, IsGroupMember=%v, want Denied", v)
	}

	// 防复活：旧版本重新加回被拒，u1 维持移除
	if got, err := c.AddGroupMember(ctx, "g1", "u1", 5); err != nil || got != SkippedStale {
		t.Fatalf("add u1@5 (stale) = (%v,%v), want SkippedStale", got, err)
	}
	if v := c.IsGroupMember(ctx, "g1", "u1"); v != VerdictDenied {
		t.Errorf("after stale re-add, IsGroupMember(u1)=%v, want Denied (no resurrection)", v)
	}

	// 防乱序：等版本号也算陈旧
	if got, _ := c.AddGroupMember(ctx, "g1", "u9", 6); got != SkippedStale {
		t.Errorf("add u9@6 (== current) = %v, want SkippedStale", got)
	}

	// 新版本加人生效
	if got, _ := c.AddGroupMember(ctx, "g1", "u3", 7); got != Applied {
		t.Errorf("add u3@7 = %v, want Applied", got)
	}
	if v := c.IsGroupMember(ctx, "g1", "u3"); v != VerdictAllowed {
		t.Errorf("after add u3@7, IsGroupMember=%v, want Allowed", v)
	}

	// 未加载的群：事件跳过维护，绝不半 populate（否则会被误判为已加载）
	if got, _ := c.RemoveGroupMember(ctx, "gX", "u1", 1); got != SkippedNotLoaded {
		t.Errorf("remove on unloaded group = %v, want SkippedNotLoaded", got)
	}
	if v := c.IsGroupMember(ctx, "gX", "u1"); v != VerdictUnknown {
		t.Errorf("unloaded group after skipped event, IsGroupMember=%v, want Unknown", v)
	}
}
