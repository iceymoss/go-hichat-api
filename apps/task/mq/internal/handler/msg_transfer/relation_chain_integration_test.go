package msg_transfer

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	mq_client "github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	pkgCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"github.com/segmentio/kafka-go"
)

const intTestBroker = "127.0.0.1:9092"
const intTestTopic = "relationChangeTransfer_inttest"

// repoConfigDir 从本测试文件回溯仓库根定位 config 目录（apps/task/mq/internal/handler/msg_transfer -> 6 级）。
func repoConfigDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..", "..", "config")
}

// TestRelationChain_Integration 端到端验证关系变更链路（真实 MySQL/Redis/Kafka，不重启任何服务、不改配置）：
//   warm 缓存 -> 同事务写 outbox -> relay ListPending -> client.Push 到 Kafka -> 读回 ->
//   ApplyRelationEvent(消费者真实派发逻辑) -> 断言 Redis grp:mem 被 SREM -> MarkSent。
// 任一基础设施不可用则 skip（遵守不 mock；测试数据用唯一前缀，结束清理）。
func TestRelationChain_Integration(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Skipf("config/mysql unavailable, skip: %v", r)
			}
		}()
		pkgCfg.InitConfig("local", "", repoConfigDir())
	}()

	ctx := context.Background()

	rdb := db.GetRedisConn()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	rc := relationcache.New(rdb)

	mysql := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
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
	om := socialmodels.NewRelationOutboxModel()
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
	client := mq_client.NewRelationChangeTransferClient([]string{intTestBroker}, intTestTopic)
	ev.Version = int64(row.ID)
	if err := client.Push(ev); err != nil {
		t.Skipf("kafka unavailable (push): %v", err)
	}

	// 5) 从 Kafka 读回本事件（按唯一 gid 匹配，跳过历史消息）
	got := readBackByGid(t, ctx, gid)
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
func readBackByGid(t *testing.T, ctx context.Context, gid string) *mq.RelationChangeTransfer {
	t.Helper()
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{intTestBroker},
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
