package relay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	relayInterval = time.Second
	relayBatch    = 200
	lockKey       = "relation:relay:lock"
	lockTTL       = 30 * time.Second
)

var relationLockRenewInterval = lockTTL / 3

// Relay 关系变更发件箱投递器：周期轮询 relation_outbox 待投递行，投递到 Kafka 后标记已投递。
// 多副本下用 Redis 单实例锁保证同一时刻只有一个 relay 在跑（mq-task.md 约定）。
type Relay struct {
	svcCtx *svc.ServiceContext
	rdb    *redis.Client
}

func New(svcCtx *svc.ServiceContext) *Relay {
	return &Relay{svcCtx: svcCtx, rdb: db.GetRedisConn()}
}

// Start 阻塞运行投递循环，直到 ctx 取消（优雅退出）。
func (r *Relay) Start(ctx context.Context) {
	ticker := time.NewTicker(relayInterval)
	defer ticker.Stop()
	zLog.Info("relation outbox relay started")
	for {
		select {
		case <-ctx.Done():
			zLog.Info("relation outbox relay stopped")
			return
		case <-ticker.C:
			r.tickOnce(ctx)
		}
	}
}

// tickOnce 抢到单实例锁后投递一批；锁被别的副本持有则本轮跳过。
func (r *Relay) tickOnce(ctx context.Context) {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return
	}
	token := hex.EncodeToString(tokenBytes)
	ok, err := r.rdb.SetNX(ctx, lockKey, token, lockTTL).Result()
	if err != nil || !ok {
		return
	}
	renewCtx, cancelRenew := context.WithCancel(ctx)
	defer cancelRenew()
	lost := make(chan struct{})
	go renewRelationLock(renewCtx, r.rdb, token, lost)
	defer r.rdb.Eval(context.Background(), `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`, []string{lockKey}, token)

	rows, err := r.svcCtx.RelationOutboxModel.ListPending(ctx, relayBatch)
	if err != nil {
		zLog.Error("relay ListPending", zap.Error(err))
		return
	}

	sent := make([]uint64, 0, len(rows))
	for _, row := range rows {
		select {
		case <-lost:
			return
		default:
		}
		var ev mq.RelationChangeTransfer
		if err := json.Unmarshal([]byte(row.Payload), &ev); err != nil {
			// 坏数据：记录并标记已投递，避免毒丸阻塞队列头（事件丢失换队列前进）
			zLog.Error("relay bad payload, skip", zap.Uint64("id", row.ID), zap.Error(err))
			sent = append(sent, row.ID)
			continue
		}
		// 用 outbox.id 回填单调版本号，保证消费端版本门一致
		ev.Version = int64(row.ID)
		if err := r.svcCtx.RelationChangeTransferClient.Push(&ev); err != nil {
			// 投递失败：停止本批，已成功的先标记，剩余下轮重试（at-least-once）
			zLog.Error("relay push failed, retry next tick", zap.Uint64("id", row.ID), zap.Error(err))
			break
		}
		sent = append(sent, row.ID)
	}

	if len(sent) > 0 {
		if err := r.svcCtx.RelationOutboxModel.MarkSent(ctx, sent); err != nil {
			zLog.Error("relay MarkSent", zap.Any("ids", sent), zap.Error(err))
		}
	}
}

func renewRelationLock(ctx context.Context, lock *redis.Client, token string, lost chan<- struct{}) {
	ticker := time.NewTicker(relationLockRenewInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := lock.Eval(ctx, `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`, []string{lockKey}, token, int64(lockTTL/time.Millisecond)).Int()
			if err != nil || result == 0 {
				close(lost)
				return
			}
		}
	}
}
