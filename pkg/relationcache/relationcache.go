package relationcache

import (
	"context"
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

const groupMemberKeyPrefix = "grp:mem:"

func groupMemberKey(gid string) string { return groupMemberKeyPrefix + gid }

// Cache 关系缓存：群成员集 grp:mem:{gid}、好友集 frd:{uid}。
// 接受注入的 *redis.Client，便于测试与在各服务 svc 中复用 db.GetRedisConn()。
type Cache struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// LoadGroupMembers 用权威来源（RPC GroupUsers 回源结果）重建群成员集。
// 始终写入 loadedSentinel，使空群也有非空 key，可与"未加载"区分。
func (c *Cache) LoadGroupMembers(ctx context.Context, gid string, members []string, version int64) error {
	key := groupMemberKey(gid)
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, key)
	args := make([]any, 0, len(members)+1)
	args = append(args, loadedSentinel)
	for _, m := range members {
		args = append(args, m)
	}
	pipe.SAdd(ctx, key, args...)
	pipe.Expire(ctx, key, cacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// IsGroupMember 三态判定 uid 是否为 gid 成员。
// 未加载 / Redis 错误 -> Unknown（调用方 fail-open）；哨兵不计为业务成员。
func (c *Cache) IsGroupMember(ctx context.Context, gid, uid string) Verdict {
	if uid == loadedSentinel {
		return VerdictDenied
	}
	key := groupMemberKey(gid)
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return VerdictUnknown
	}
	if exists == 0 {
		return VerdictUnknown
	}
	isMember, err := c.rdb.SIsMember(ctx, key, uid).Result()
	if err != nil {
		return VerdictUnknown
	}
	if isMember {
		return VerdictAllowed
	}
	return VerdictDenied
}
