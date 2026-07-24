package socialmodels

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	pkgCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
)

// repoConfigDir 通过本测试文件位置回溯仓库根，定位 config 目录（不依赖测试 CWD）。
func repoConfigDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// apps/social/socialmodels/ -> 回溯三级到仓库根
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "config")
}

// newOutboxModelForTest 初始化配置 + 真实 MySQL 连接；不可用则 skip（遵守不 mock DB）。
func newOutboxModelForTest(t *testing.T) RelationOutboxModel {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("config/mysql unavailable, skip: %v", r)
		}
	}()
	pkgCfg.InitConfig("local", repoConfigDir())
	conn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	if err := conn.Exec("SELECT 1").Error; err != nil {
		t.Skipf("mysql unavailable, skip: %v", err)
	}
	return NewRelationOutboxModel()
}

func Test_RelationOutbox_InsertListMarkSent(t *testing.T) {
	m := newOutboxModelForTest(t)
	ctx := context.Background()
	conn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	table := objects.RelationOutbox{}.TableName()

	// 用唯一 event_type 标记本次测试的行，测试后清理
	marker := "test.relation.outbox." + time.Now().Format("150405.000000")
	t.Cleanup(func() {
		conn.Table(table).Where("event_type = ?", marker).Delete(nil)
	})

	// InsertTx：在一个事务里写入两条
	tx := conn.Begin()
	for _, gid := range []string{"g100", "g101"} {
		row := &objects.RelationOutbox{EventType: marker, GroupID: gid, Payload: `{"gid":"` + gid + `"}`, Status: 0}
		if err := m.InsertTx(tx, row); err != nil {
			tx.Rollback()
			t.Fatalf("InsertTx: %v", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	// ListPending：能取到本次插入的待投递行，按 id 升序
	rows, err := m.ListPending(ctx, 100)
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	var mine []*objects.RelationOutbox
	for _, r := range rows {
		if r.EventType == marker {
			mine = append(mine, r)
		}
	}
	if len(mine) != 2 {
		t.Fatalf("ListPending found %d marked pending, want 2", len(mine))
	}
	if mine[0].ID >= mine[1].ID {
		t.Errorf("pending not ordered by id asc: %d, %d", mine[0].ID, mine[1].ID)
	}

	// MarkSent：标记第一条已投递后，ListPending 不应再返回它
	if err := m.MarkSent(ctx, []uint64{mine[0].ID}); err != nil {
		t.Fatalf("MarkSent: %v", err)
	}
	rows2, _ := m.ListPending(ctx, 100)
	for _, r := range rows2 {
		if r.ID == mine[0].ID {
			t.Errorf("MarkSent row still pending: id=%d", r.ID)
		}
	}
}
