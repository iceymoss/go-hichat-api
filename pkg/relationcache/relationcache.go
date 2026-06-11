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

const groupMemberKeyPrefix = "grp:mem:"
const friendKeyPrefix = "frd:"

func groupMemberKey(gid string) string { return groupMemberKeyPrefix + gid }
func groupVerKey(gid string) string    { return groupMemberKeyPrefix + gid + ":ver" }
func groupLockKey(gid string) string   { return groupMemberKeyPrefix + gid + ":lock" }
func friendSetKey(uid string) string   { return friendKeyPrefix + uid }
func friendVerKey(uid string) string   { return friendKeyPrefix + uid + ":ver" }

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

// applyMemberScript 原子地按版本门维护集合成员：
//   未加载(set 不存在) -> 返回 -1；version<=当前 -> 返回 2；否则 SADD/SREM + 更新版本 + 续期 -> 返回 1。
// 用 Lua 保证"比较版本 + 改集合 + 写版本"三步原子，杜绝并发下的乱序与读穿透复活竞态。
var applyMemberScript = redis.NewScript(`
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

// LoadGroupMembers 用权威来源（RPC GroupUsers 回源结果）重建群成员集，并落版本号。
func (c *Cache) LoadGroupMembers(ctx context.Context, gid string, members []string, version int64) error {
	return c.loadSet(ctx, groupMemberKey(gid), groupVerKey(gid), members, version)
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

// AddGroupMember 版本门下把 uid 加入群成员集（加群/邀请事件）。
func (c *Cache) AddGroupMember(ctx context.Context, gid, uid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "add", groupMemberKey(gid), groupVerKey(gid), uid, version)
}

// RemoveGroupMember 版本门下把 uid 移出群成员集（踢人/退群事件）。
func (c *Cache) RemoveGroupMember(ctx context.Context, gid, uid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "rem", groupMemberKey(gid), groupVerKey(gid), uid, version)
}

// LoadFriends 用权威来源（RPC FriendList 回源结果）重建 uid 的好友集，并落版本号。
func (c *Cache) LoadFriends(ctx context.Context, uid string, friends []string, version int64) error {
	return c.loadSet(ctx, friendSetKey(uid), friendVerKey(uid), friends, version)
}

// IsFriend 三态判定 friendUid 是否在 uid 的好友集内。
func (c *Cache) IsFriend(ctx context.Context, uid, friendUid string) Verdict {
	return c.verdict(ctx, friendSetKey(uid), friendUid)
}

// AddFriend 版本门下把 friendUid 加入 uid 的好友集（加好友事件，双向各调一次）。
func (c *Cache) AddFriend(ctx context.Context, uid, friendUid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "add", friendSetKey(uid), friendVerKey(uid), friendUid, version)
}

// RemoveFriend 版本门下把 friendUid 移出 uid 的好友集（删好友事件，双向各调一次）。
func (c *Cache) RemoveFriend(ctx context.Context, uid, friendUid string, version int64) (ApplyResult, error) {
	return c.applyMember(ctx, "rem", friendSetKey(uid), friendVerKey(uid), friendUid, version)
}

// loadSet 用权威快照重建一个成员集合（始终含 loadedSentinel），并落版本号。
func (c *Cache) loadSet(ctx context.Context, setKey, verKey string, members []string, version int64) error {
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, setKey)
	args := make([]any, 0, len(members)+1)
	args = append(args, loadedSentinel)
	for _, m := range members {
		args = append(args, m)
	}
	pipe.SAdd(ctx, setKey, args...)
	pipe.Set(ctx, verKey, version, cacheTTL)
	pipe.Expire(ctx, setKey, cacheTTL)
	_, err := pipe.Exec(ctx)
	return err
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

func (c *Cache) applyMember(ctx context.Context, op, setKey, verKey, uid string, version int64) (ApplyResult, error) {
	ttlMs := cacheTTL.Milliseconds()
	res, err := applyMemberScript.Run(ctx, c.rdb, []string{setKey, verKey}, op, uid, version, ttlMs).Int64()
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
	return c.rdb.Del(ctx, groupMemberKey(gid), groupVerKey(gid)).Err()
}

// InvalidateFriend 删除某用户的好友集缓存。
func (c *Cache) InvalidateFriend(ctx context.Context, uid string) error {
	return c.rdb.Del(ctx, friendSetKey(uid), friendVerKey(uid)).Err()
}

// sremIfLoadedScript 集合已加载则无条件 SREM 并续期；未加载返回 -1（跳过）。
var sremIfLoadedScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then return -1 end
redis.call('SREM', KEYS[1], ARGV[1])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return 1
`)

// RemoveFriendIfLoaded 无版本门移除好友。
// 好友事件按好友对分区——同一用户的不同好友变更事件跨分区到达，
// 不能套用单一 per-user 版本门（否则会误拒较小版本号的删除，导致漏删残留）。删除可交换、重放安全，故无条件 SREM。
func (c *Cache) RemoveFriendIfLoaded(ctx context.Context, uid, friendUid string) (ApplyResult, error) {
	res, err := sremIfLoadedScript.Run(ctx, c.rdb, []string{friendSetKey(uid)}, friendUid, cacheTTL.Milliseconds()).Int64()
	if err != nil {
		return SkippedNotLoaded, err
	}
	if res == 1 {
		return Applied, nil
	}
	return SkippedNotLoaded, nil
}

// saddIfLoadedScript 集合已加载则无条件 SADD 并续期；未加载返回 -1（跳过，留给读穿透重建）。
var saddIfLoadedScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then return -1 end
redis.call('SADD', KEYS[1], ARGV[1])
redis.call('PEXPIRE', KEYS[1], ARGV[2])
return 1
`)

// AddFriendIfLoaded 无版本门加好友（与 RemoveFriendIfLoaded 对称）。
// 好友新增事件按好友对分区：与同一好友对的删除事件落同分区、天然有序（不会复活已删关系）；
// 而同一用户的不同好友对事件跨分区，无法套用单一 per-user 版本门。SADD 幂等可交换、重放安全，故无条件 SADD。
// 未加载则跳过：绝不半 populate，留给读穿透从权威 FriendList 重建。
func (c *Cache) AddFriendIfLoaded(ctx context.Context, uid, friendUid string) (ApplyResult, error) {
	res, err := saddIfLoadedScript.Run(ctx, c.rdb, []string{friendSetKey(uid)}, friendUid, cacheTTL.Milliseconds()).Int64()
	if err != nil {
		return SkippedNotLoaded, err
	}
	if res == 1 {
		return Applied, nil
	}
	return SkippedNotLoaded, nil
}
