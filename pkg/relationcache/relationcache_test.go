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

func Test_Friend_VerdictAndVersionGate(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()

	// uid=a 的好友集 {b, c}
	if err := c.LoadFriends(ctx, "a", []string{"b", "c"}, 3); err != nil {
		t.Fatalf("load friends a: %v", err)
	}

	tests := []struct {
		name   string
		uid    string
		friend string
		want   Verdict
	}{
		{"friend -> allowed", "a", "b", VerdictAllowed},
		{"non-friend, loaded -> denied", "a", "z", VerdictDenied},
		{"not loaded -> unknown (fail-open)", "x", "y", VerdictUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.IsFriend(ctx, tt.uid, tt.friend); got != tt.want {
				t.Errorf("IsFriend(%s,%s) = %v, want %v", tt.uid, tt.friend, got, tt.want)
			}
		})
	}

	// 删好友：新版本生效
	if got, _ := c.RemoveFriend(ctx, "a", "b", 4); got != Applied {
		t.Errorf("remove friend b@4 = %v, want Applied", got)
	}
	if v := c.IsFriend(ctx, "a", "b"); v != VerdictDenied {
		t.Errorf("after remove friend b, IsFriend=%v, want Denied", v)
	}
	// 旧版本不能复活已删好友
	if got, _ := c.AddFriend(ctx, "a", "b", 3); got != SkippedStale {
		t.Errorf("stale add friend b@3 = %v, want SkippedStale", got)
	}
	if v := c.IsFriend(ctx, "a", "b"); v != VerdictDenied {
		t.Errorf("after stale re-add friend, IsFriend=%v, want Denied", v)
	}
}

func Test_GroupMembers_FanoutRead(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()

	if err := c.LoadGroupMembers(ctx, "g1", []string{"u1", "u2"}, 1); err != nil {
		t.Fatalf("load g1: %v", err)
	}
	if err := c.LoadGroupMembers(ctx, "g2", []string{}, 1); err != nil {
		t.Fatalf("load g2: %v", err)
	}

	// 未加载：loaded=false，调用方据此回源 RPC
	if members, loaded, err := c.GroupMembers(ctx, "gX"); err != nil || loaded || len(members) != 0 {
		t.Errorf("GroupMembers(unloaded) = (%v,%v,%v), want ([],false,nil)", members, loaded, err)
	}

	// 已加载非空：返回真实成员，剔除哨兵
	members, loaded, err := c.GroupMembers(ctx, "g1")
	if err != nil || !loaded {
		t.Fatalf("GroupMembers(g1) = (_,%v,%v), want loaded,nil", loaded, err)
	}
	got := map[string]bool{}
	for _, m := range members {
		got[m] = true
	}
	if len(members) != 2 || !got["u1"] || !got["u2"] || got[loadedSentinel] {
		t.Errorf("GroupMembers(g1) = %v, want exactly {u1,u2} without sentinel", members)
	}

	// 已加载空群：loaded=true，成员为空（哨兵剔除）
	if members, loaded, err := c.GroupMembers(ctx, "g2"); err != nil || !loaded || len(members) != 0 {
		t.Errorf("GroupMembers(empty g2) = (%v,%v,%v), want ([],true,nil)", members, loaded, err)
	}
}

func Test_SingleFlightLock(t *testing.T) {
	c, cleanup := newTestCache(t)
	defer cleanup()
	ctx := context.Background()

	// 第一个抢锁成功
	token, ok, err := c.TryLockGroupRebuild(ctx, "g1")
	if err != nil || !ok || token == "" {
		t.Fatalf("first lock = (%q,%v,%v), want acquired", token, ok, err)
	}

	// 持锁期间第二个抢锁失败（防击穿：只放一个回源）
	if _, ok2, _ := c.TryLockGroupRebuild(ctx, "g1"); ok2 {
		t.Errorf("second lock acquired while held, want blocked")
	}

	// 释放后可再次抢到
	c.UnlockGroupRebuild(ctx, "g1", token)
	if _, ok3, _ := c.TryLockGroupRebuild(ctx, "g1"); !ok3 {
		t.Errorf("lock after unlock not acquired, want acquired")
	}

	// 用错 token 不能误删他人锁
	other, _, _ := c.TryLockGroupRebuild(ctx, "g2")
	c.UnlockGroupRebuild(ctx, "g2", "wrong-token")
	if _, ok4, _ := c.TryLockGroupRebuild(ctx, "g2"); ok4 {
		t.Errorf("lock released by wrong token, want still held")
	}
	c.UnlockGroupRebuild(ctx, "g2", other)
}
