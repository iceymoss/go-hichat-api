package relationcache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/go-redis/redis/v8"
)

// Verdict 是关系判定的三态结果。Unknown 表示"无法确定"（缓存未加载 / Redis 错误），
// 调用方据此 fail-open（放行）；只有 Denied 才是"确定不是成员/好友"，调用方才拒绝。
type Verdict int

const (
	VerdictUnknown Verdict = iota // 不确定 -> 调用方 fail-open 放行
	VerdictAllowed                // 确定在集合内
	VerdictDenied                 // 确定不在集合内（集合已加载但目标缺席）
)

// loadedSentinel 占位成员：保证"已加载但为空"的集合 key 仍存在，
// 从而把"空群（Denied）"与"未加载（Unknown）"区分开。业务 uid 永不等于该值。
const loadedSentinel = "__loaded__"

// cacheTTL 缓存 key 的存活时间，作为事件漏更时的最终自愈上界（spec L4）。
const cacheTTL = time.Hour

// rebuildLockTTL 冷启动单飞锁的存活时间，回源 RPC 应在此之内完成；过期自动释放防死锁。
const rebuildLockTTL = 5 * time.Second

// tombstoneTTL 墓碑存活时间：标记"刚被移除"的成员，让冷缓存读穿透回填时剔除之，
// 杜绝"读穿透 RPC 在移除提交前取到旧快照、提交后才回填"的 TOCTOU 复活。
// 取值需覆盖读穿透 RPC（≤2s）+ relay/消费滞后；同群/同好友对事件有序，re-add 会主动清墓碑，TTL 仅作安全上界。
const tombstoneTTL = 60 * time.Second

const groupMemberKeyPrefix = "grp:mem:"
const friendKeyPrefix = "frd:"

func groupMemberKey(gid string) string { return groupMemberKeyPrefix + gid }
func groupVerKey(gid string) string    { return groupMemberKeyPrefix + gid + ":ver" }
func groupLockKey(gid string) string   { return groupMemberKeyPrefix + gid + ":lock" }
func groupTombKey(gid string) string   { return groupMemberKeyPrefix + gid + ":tomb" }
func friendSetKey(uid string) string   { return friendKeyPrefix + uid }
func friendVerKey(uid string) string   { return friendKeyPrefix + uid + ":ver" }
func friendTombKey(uid string) string  { return friendKeyPrefix + uid + ":tomb" }

// unlockScript 仅当锁值等于自己持有的 token 时才删除，避免误删他人（或锁过期后被他人重抢）的锁。
var unlockScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// ApplyResult 是版本门维护操作（事件驱动 SADD/SREM）的结果。
type ApplyResult int

const (
	Applied          ApplyResult = iota // 已应用
	SkippedStale                        // version <= 当前版本，按乱序/重放丢弃（含防"复活"）
	SkippedNotLoaded                    // 集合未加载：跳过，绝不半 populate（留给读穿透重建）
)

// applyMemberScript 原子地按版本门维护集合成员 + 无条件维护墓碑：
//   墓碑维护始终执行（rem→SADD 墓碑+续期；add→SREM 墓碑），不受集合是否加载/版本门影响，
//   保证冷缓存下的移除也能被随后的读穿透剔除、re-add 能清掉墓碑让成员被重新纳入。
//   集合改动仍走版本门：未加载(set 不存在) -> 返回 -1；version<=当前 -> 返回 2；否则 SADD/SREM + 更新版本 + 续期 -> 返回 1。
// 用 Lua 保证"墓碑 + 比较版本 + 改集合 + 写版本"原子，杜绝并发下的乱序与读穿透复活竞态。
var applyMemberScript = redis.NewScript(`
if ARGV[1] == 'add' then
  redis.call('SREM', KEYS[3], ARGV[2])
else
  redis.call('SADD', KEYS[3], ARGV[2])
  redis.call('PEXPIRE', KEYS[3], ARGV[5])
end
if redis.call('EXISTS', KEYS[1]) == 0 then return -1 end
local cur = tonumber(redis.call('GET', KEYS[2]) or '-1')
local v = tonumber(ARGV[3])
if v <= cur then return 2 end
if ARGV[1] == 'add' then
  redis.call('SADD', KEYS[1], ARGV[2])
else
  redis.call('SREM', KEYS[1], ARGV[2])
end
redis.call('SET', KEYS[2], v)
redis.call('PEXPIRE', KEYS[1], ARGV[4])
redis.call('PEXPIRE', KEYS[2], ARGV[4])
return 1
`)

// Cache 关系缓存：群成员集 grp:mem:{gid}、好友集 frd:{uid}。
// 接受注入的 *redis.Client，便于测试与在各服务 svc 中复用 db.GetRedisConn()。
type Cache struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// LoadGroupMembers 用权威来源（RPC GroupUsers 回源结果）重建群成员集，并落版本号。仅冷缓存时生效。
func (c *Cache) LoadGroupMembers(ctx context.Context, gid string, members []string, version int64) error {
	return c.loadSet(ctx, groupMemberKey(gid), groupVerKey(gid), groupTombKey(gid), members, version)
}

// IsGroupMember 三态判定 uid 是否为 gid 成员。
func (c *Cache) IsGroupMember(ctx context.Context, gid, uid string) Verdict {
	return c.verdict(ctx, groupMemberKey(gid), uid)
}

// GroupMembers 返回群成员列表（剔除哨兵）供扇出使用。
// loaded=false 表示集合未加载，调用方应回源 RPC 后 LoadGroupMembers，再用本次 RPC 结果扇出。
func (c *Cache) GroupMembers(ctx context.Context, gid string) (members []string, loaded bool, err error) {
	key := groupMemberKey(gid)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}
	raw, err := c.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	members = make([]string, 0, len(raw))
	for _, m := range raw {
		if m == loadedSentinel {
			continue
		}
		members = append(members, m)
	}
	return members, true, nil
}

// AddGroupMember 版本门下把 uid 加入群成员集（加群/邀请事件），并清掉其墓碑（允许 re-add）。
func (c *Cache) AddGroupMember(ctx context.Context, gid, uid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "add", groupMemberKey(gid), groupVerKey(gid), groupTombKey(gid), uid, version)
}

// RemoveGroupMember 版本门下把 uid 移出群成员集（踢人/退群事件），并无条件写墓碑（防读穿透复活）。
func (c *Cache) RemoveGroupMember(ctx context.Context, gid, uid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "rem", groupMemberKey(gid), groupVerKey(gid), groupTombKey(gid), uid, version)
}

// LoadFriends 用权威来源（RPC FriendList 回源结果）重建 uid 的好友集，并落版本号。仅冷缓存时生效。
func (c *Cache) LoadFriends(ctx context.Context, uid string, friends []string, version int64) error {
	return c.loadSet(ctx, friendSetKey(uid), friendVerKey(uid), friendTombKey(uid), friends, version)
}

// IsFriend 三态判定 friendUid 是否在 uid 的好友集内。
func (c *Cache) IsFriend(ctx context.Context, uid, friendUid string) Verdict {
	return c.verdict(ctx, friendSetKey(uid), friendUid)
}

// AddFriend 版本门下把 friendUid 加入 uid 的好友集（加好友事件，双向各调一次）。
func (c *Cache) AddFriend(ctx context.Context, uid, friendUid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "add", friendSetKey(uid), friendVerKey(uid), friendTombKey(uid), friendUid, version)
}

// RemoveFriend 版本门下把 friendUid 移出 uid 的好友集（删好友事件，双向各调一次）。
func (c *Cache) RemoveFriend(ctx context.Context, uid, friendUid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "rem", friendSetKey(uid), friendVerKey(uid), friendTombKey(uid), friendUid, version)
}

// loadIfColdScript 仅在集合"冷"（不存在）时用权威快照重建，热缓存直接返回 0 不覆盖——
// 保证缓存一旦由读穿透/事件加载，后续就只由版本门事件维护，杜绝滞后读穿透把 ver 回退、复活已移除成员。
// 重建时逐个剔除墓碑命中的成员：覆盖"读穿透 RPC 在移除提交前取到旧快照、提交后才回填"的 TOCTOU。
//   KEYS[1]=set KEYS[2]=ver KEYS[3]=tomb；ARGV[1]=version ARGV[2]=ttlMs ARGV[3]=sentinel ARGV[4..]=members。
//   热 -> 0；冷重建 -> 1。
var loadIfColdScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then return 0 end
redis.call('DEL', KEYS[1])
redis.call('SADD', KEYS[1], ARGV[3])
for i=4,#ARGV do
  if redis.call('SISMEMBER', KEYS[3], ARGV[i]) == 0 then
    redis.call('SADD', KEYS[1], ARGV[i])
  end
end
redis.call('SET', KEYS[2], ARGV[1])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
redis.call('PEXPIRE', KEYS[2], ARGV[2])
return 1
`)

// loadSet 仅在冷缓存时用权威快照重建一个成员集合（始终含 loadedSentinel、剔除墓碑成员），并落版本号。
// 热缓存为 no-op：read-through 只负责暖冷缓存，绝不覆盖已被事件维护的热缓存。
func (c *Cache) loadSet(ctx context.Context, setKey, verKey, tombKey string, members []string, version int64) error {
	ttlMs := cacheTTL.Milliseconds()
	args := make([]any, 0, len(members)+3)
	args = append(args, version, ttlMs, loadedSentinel)
	for _, m := range members {
		args = append(args, m)
	}
	return loadIfColdScript.Run(ctx, c.rdb, []string{setKey, verKey, tombKey}, args...).Err()
}

// verdict 三态判定 member 是否在 setKey 集合内。
// 未加载 / Redis 错误 -> Unknown（调用方 fail-open）；哨兵不计为业务成员。
func (c *Cache) verdict(ctx context.Context, setKey, member string) Verdict {
	if member == loadedSentinel {
		return VerdictDenied
	}
	exists, err := c.rdb.Exists(ctx, setKey).Result()
	if err != nil {
		return VerdictUnknown
	}
	if exists == 0 {
		return VerdictUnknown
	}
	isMember, err := c.rdb.SIsMember(ctx, setKey, member).Result()
	if err != nil {
		return VerdictUnknown
	}
	if isMember {
		return VerdictAllowed
	}
	return VerdictDenied
}

func (c *Cache) applyMember(ctx context.Context, op, setKey, verKey, tombKey, uid string, version int64) (ApplyResult, error) {
	ttlMs := cacheTTL.Milliseconds()
	tombTtlMs := tombstoneTTL.Milliseconds()
	res, err := applyMemberScript.Run(ctx, c.rdb, []string{setKey, verKey, tombKey}, op, uid, version, ttlMs, tombTtlMs).Int64()
	if err != nil {
		return SkippedNotLoaded, err
	}
	switch res {
	case 1:
		return Applied, nil
	case 2:
		return SkippedStale, nil
	default: // -1
		return SkippedNotLoaded, nil
	}
}

// TryLockGroupRebuild 抢占某群的冷启动重建锁，防缓存击穿（同一群同一时刻只放一个回源 RPC）。
// 返回的 token 用于安全释放；ok=false 表示已有别人在重建，调用方应短暂等待或本条降级直连 RPC。
func (c *Cache) TryLockGroupRebuild(ctx context.Context, gid string) (token string, ok bool, err error) {
	token, err = randToken()
	if err != nil {
		return "", false, err
	}
	ok, err = c.rdb.SetNX(ctx, groupLockKey(gid), token, rebuildLockTTL).Result()
	if err != nil || !ok {
		return "", false, err
	}
	return token, true, nil
}

// UnlockGroupRebuild 释放重建锁；仅当持有 token 匹配时才删除，best-effort 忽略错误。
func (c *Cache) UnlockGroupRebuild(ctx context.Context, gid, token string) {
	if token == "" {
		return
	}
	_ = unlockScript.Run(ctx, c.rdb, []string{groupLockKey(gid)}, token).Err()
}

func randToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// InvalidateGroup 删除整群缓存（群解散事件）。删除后判定退化为 Unknown（fail-open），
// 下次读穿透回源；解散群本身已无会话，无需重建。best-effort 忽略错误。
func (c *Cache) InvalidateGroup(ctx context.Context, gid string) error {
	return c.rdb.Del(ctx, groupMemberKey(gid), groupVerKey(gid), groupTombKey(gid)).Err()
}

// InvalidateFriend 删除某用户的好友集缓存。
func (c *Cache) InvalidateFriend(ctx context.Context, uid string) error {
	return c.rdb.Del(ctx, friendSetKey(uid), friendVerKey(uid), friendTombKey(uid)).Err()
}

// sremIfLoadedScript 无条件写墓碑（防读穿透复活）；集合已加载则 SREM 并续期；未加载返回 -1（跳过）。
//   KEYS[1]=set KEYS[2]=tomb；ARGV[1]=member ARGV[2]=ttlMs ARGV[3]=tombTtlMs。
var sremIfLoadedScript = redis.NewScript(`
redis.call('SADD', KEYS[2], ARGV[1])
redis.call('PEXPIRE', KEYS[2], ARGV[3])
if redis.call('EXISTS', KEYS[1]) == 0 then return -1 end
redis.call('SREM', KEYS[1], ARGV[1])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return 1
`)

// RemoveFriendIfLoaded 无版本门移除好友（无条件写墓碑）。
// 好友事件按好友对分区——同一用户的不同好友变更事件跨分区到达，
// 不能套用单一 per-user 版本门（否则会误拒较小版本号的删除，导致漏删残留）。删除可交换、重放安全，故无条件 SREM。
func (c *Cache) RemoveFriendIfLoaded(ctx context.Context, uid, friendUid string) (ApplyResult, error) {
	res, err := sremIfLoadedScript.Run(ctx, c.rdb, []string{friendSetKey(uid), friendTombKey(uid)},
		friendUid, cacheTTL.Milliseconds(), tombstoneTTL.Milliseconds()).Int64()
	if err != nil {
		return SkippedNotLoaded, err
	}
	if res == 1 {
		return Applied, nil
	}
	return SkippedNotLoaded, nil
}

// saddIfLoadedScript 无条件清墓碑（允许 re-add）；集合已加载则 SADD 并续期；未加载返回 -1（跳过，留给读穿透重建）。
//   KEYS[1]=set KEYS[2]=tomb；ARGV[1]=member ARGV[2]=ttlMs。
var saddIfLoadedScript = redis.NewScript(`
redis.call('SREM', KEYS[2], ARGV[1])
if redis.call('EXISTS', KEYS[1]) == 0 then return -1 end
redis.call('SADD', KEYS[1], ARGV[1])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return 1
`)

// AddFriendIfLoaded 无版本门加好友（无条件清墓碑，与 RemoveFriendIfLoaded 对称）。
// 好友新增事件按好友对分区：与同一好友对的删除事件落同分区、天然有序（不会复活已删关系）；
// 而同一用户的不同好友对事件跨分区，无法套用单一 per-user 版本门。SADD 幂等可交换、重放安全，故无条件 SADD。
// 未加载则跳过：绝不半 populate，留给读穿透从权威 FriendList 重建（墓碑已清，重建会纳入该好友）。
func (c *Cache) AddFriendIfLoaded(ctx context.Context, uid, friendUid string) (ApplyResult, error) {
	res, err := saddIfLoadedScript.Run(ctx, c.rdb, []string{friendSetKey(uid), friendTombKey(uid)},
		friendUid, cacheTTL.Milliseconds()).Int64()
	if err != nil {
		return SkippedNotLoaded, err
	}
	if res == 1 {
		return Applied, nil
	}
	return SkippedNotLoaded, nil
}
