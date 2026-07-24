package msg_transfer

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	mq_client "github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const intTestTopic = "relationChangeTransfer_inttest"

// TestRelationChain_Integration 端到端验证关系变更链路（真实 MySQL/Redis/Kafka，不重启任何服务、不改配置）：
//
//	warm 缓存 -> 同事务写 outbox -> relay ListPending -> client.Push 到 Kafka -> 读回 ->
//	ApplyRelationEvent(消费者真实派发逻辑) -> 断言 Redis grp:mem 被 SREM -> MarkSent。
//
// 任一基础设施不可用则 skip（遵守不 mock；测试数据用唯一前缀，结束清理）。
func TestRelationChain_Integration(t *testing.T) {
	broker := os.Getenv("HICHAT_RELATION_CHAIN_KAFKA_BROKER")
	if broker == "" {
		broker = "127.0.0.1:9092"
	}
	mysqlDSN := os.Getenv("HICHAT_RELATION_CHAIN_MYSQL_DSN")
	redisAddr := os.Getenv("HICHAT_RELATION_CHAIN_REDIS_ADDR")
	if mysqlDSN == "" || redisAddr == "" || os.Getenv("HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS") != "1" {
		t.Skip("set explicit relation-chain dependencies and HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS=1 to run")
	}
	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("open integration MySQL: %v", err)
	}
	var databaseName string
	if err := mysqlDB.Raw("SELECT DATABASE()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("read integration database: %v", err)
	}
	matched, err := regexp.MatchString(`^hichat_[a-z0-9_]*_test$`, strings.ToLower(databaseName))
	if err != nil || !matched {
		t.Fatalf("relation integration requires a dedicated hichat_*_test database, got %q", databaseName)
	}
	rdb := redisv8.NewClient(&redisv8.Options{Addr: redisAddr})
	t.Cleanup(func() { _ = rdb.Close() })

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	rc := relationcache.New(rdb)

	mysql := mysqlDB
	if err := mysql.Exec("SELECT 1").Error; err != nil {
		t.Skipf("mysql unavailable: %v", err)
	}

	uniq := strconv.FormatInt(time.Now().UnixNano(), 10)
	gid := "inttest_g_" + uniq
	u1, u2 := "inttest_u1_"+uniq, "inttest_u2_"+uniq

	t.Cleanup(func() {
		rdb.Del(context.Background(), "grp:mem:"+gid, "grp:mem:"+gid+":ver")
		mysql.Table(objects.RelationOutbox{}.TableName()).Where("group_id = ?", gid).Delete(nil)
	})

	// 1) 预热群成员缓存（模拟此前扇出已加载），断言 u2 在群内
	if err := rc.LoadGroupMembers(ctx, gid, []string{u1, u2}, 0); err != nil {
		t.Fatalf("LoadGroupMembers: %v", err)
	}
	if v := rc.IsGroupMember(ctx, gid, u2); v != relationcache.VerdictAllowed {
		t.Fatalf("precondition: IsGroupMember(u2)=%v, want Allowed", v)
	}

	// 2) social 侧：同事务写 outbox（这里直接用 model 模拟 emitRelationChange 的 outbox 写入）
	om := socialmodels.NewRelationOutboxModelWithDB(mysql)
	ev := &mq.RelationChangeTransfer{
		EventType:  constants.RelationEventGroupMemberRemoved,
		GroupId:    gid,
		UserId:     u2,
		OperatorId: u1,
		Timestamp:  time.Now().UnixMilli(),
	}
	body, _ := json.Marshal(ev)
	row := &objects.RelationOutbox{EventType: ev.EventType, GroupID: gid, Payload: string(body), Status: 0}
	tx := mysql.Begin()
	if err := om.InsertTx(tx, row); err != nil {
		tx.Rollback()
		t.Fatalf("InsertTx outbox: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}
	if row.ID == 0 {
		t.Fatalf("outbox id not assigned")
	}

	// 3) relay 步骤：ListPending 能取到本行
	pending, err := om.ListPending(ctx, 500)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if !containsOutbox(pending, row.ID) {
		t.Fatalf("ListPending did not return outbox id=%d", row.ID)
	}

	// 4) relay 步骤：用 outbox.id 作版本号投递到 Kafka
	client := mq_client.NewRelationChangeTransferClient([]string{broker}, intTestTopic)
	ev.Version = int64(row.ID)
	if err := client.Push(ev); err != nil {
		t.Skipf("kafka unavailable (push): %v", err)
	}

	// 5) 从 Kafka 读回本事件（按唯一 gid 匹配，跳过历史消息）
	got := readBackByGid(t, ctx, broker, gid)
	if got == nil {
		t.Skip("kafka round-trip: own message not received in time (kafka slow/unavailable)")
	}
	if got.Version != int64(row.ID) || got.UserId != u2 {
		t.Fatalf("round-trip mismatch: got version=%d user=%s, want version=%d user=%s", got.Version, got.UserId, row.ID, u2)
	}

	// 6) 消费者真实派发逻辑应用到缓存
	ApplyRelationEvent(ctx, rc, got)

	// 7) 断言：被踢者 u2 判定为 Denied，留存的 u1 仍 Allowed
	if v := rc.IsGroupMember(ctx, gid, u2); v != relationcache.VerdictDenied {
		t.Errorf("after event, IsGroupMember(u2)=%v, want Denied (cache not SREM'd)", v)
	}
	if v := rc.IsGroupMember(ctx, gid, u1); v != relationcache.VerdictAllowed {
		t.Errorf("after event, IsGroupMember(u1)=%v, want Allowed", v)
	}

	// 8) relay 步骤：MarkSent 后不再 pending
	if err := om.MarkSent(ctx, []uint64{row.ID}); err != nil {
		t.Fatalf("MarkSent: %v", err)
	}
	pending2, _ := om.ListPending(ctx, 500)
	if containsOutbox(pending2, row.ID) {
		t.Errorf("outbox id=%d still pending after MarkSent", row.ID)
	}
}

func containsOutbox(rows []*objects.RelationOutbox, id uint64) bool {
	for _, r := range rows {
		if r.ID == id {
			return true
		}
	}
	return false
}

// readBackByGid 从 intTestTopic 读消息，返回第一条 groupId 匹配的事件；超时返回 nil。
func readBackByGid(t *testing.T, ctx context.Context, broker, gid string) *mq.RelationChangeTransfer {
	t.Helper()
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       intTestTopic,
		GroupID:     "inttest_reader_" + gid,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	readCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	for {
		m, err := reader.ReadMessage(readCtx)
		if err != nil {
			return nil
		}
		var ev mq.RelationChangeTransfer
		if json.Unmarshal(m.Value, &ev) != nil {
			continue
		}
		if ev.GroupId == gid {
			return &ev
		}
	}
}
