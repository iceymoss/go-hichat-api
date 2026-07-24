package model

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// newNotificationModelForTest 初始化配置 + 真实 MySQL 连接，并 AutoMigrate 通知表；
// 不可用则 skip（遵守不 mock DB）。AutoMigrate 跨 SQLite/MySQL/PostgreSQL 通用，顺带校验三库建表。
func newNotificationModelForTest(t *testing.T) (NotificationModel, *gorm.DB) {
	t.Helper()
	if dsn := os.Getenv("HICHAT_NOTIFICATION_TEST_MYSQL_DSN"); dsn != "" {
		if os.Getenv("HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS") != "1" {
			t.Fatal("set HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS=1 for external database tests")
		}
		conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Fatalf("open notification test MySQL: %v", err)
		}
		var databaseName string
		if err := conn.Raw("SELECT DATABASE()").Scan(&databaseName).Error; err != nil {
			t.Fatalf("read notification test database: %v", err)
		}
		matched, err := regexp.MatchString(`^hichat_[a-z0-9_]*_test$`, strings.ToLower(databaseName))
		if err != nil || !matched {
			t.Fatalf("notification integration requires a dedicated hichat_*_test database, got %q", databaseName)
		}
		if err := conn.AutoMigrate(&objects.Notification{}, &objects.NotificationReadIntent{}); err != nil {
			t.Fatalf("migrate notifications: %v", err)
		}
		return &customNotificationModel{table: "notifications", connDB: conn}, conn
	}
	t.Skip("set HICHAT_NOTIFICATION_TEST_MYSQL_DSN and HICHAT_ALLOW_DESTRUCTIVE_DB_TESTS=1 to run")
	return nil, nil
}

// Test_Notification_Insert_Idempotent_MultiReceiver 守护：
//  1. 幂等：同一 (receiver, type, biz) 重复 Insert 只落一行，第二次 inserted=false；
//  2. 多接收者：同 type+biz 不同 receiver 各自独立成行（group.apply 扇出语义）；
//  3. 列表/未读数/标读：ListByReceiver(unreadOnly)、CountUnread、MarkRead(部分/全部) 行为正确。
func Test_Notification_Insert_Idempotent_MultiReceiver(t *testing.T) {
	m, conn := newNotificationModelForTest(t)
	ctx := context.Background()
	table := objects.Notification{}.TableName()

	// 用唯一 notify_type 标记本次测试的行，测试后清理
	marker := "test.notify." + time.Now().Format("150405.000000")
	receiverA := time.Now().Format("150405000000001")
	receiverB := time.Now().Format("150405000000002")
	t.Cleanup(func() {
		conn.Table(table).Where("notify_type = ?", marker).Delete(nil)
		conn.Table(objects.NotificationReadIntent{}.TableName()).Where("notify_type = ?", marker).Delete(nil)
	})

	newRow := func(receiver, biz string) *Notification {
		return &Notification{ReceiverId: receiver, NotifyType: marker, BizId: biz, ActorId: "actor1", Content: "hi"}
	}

	// 1. 幂等：首次 true，重复 false
	if ins, err := m.Insert(ctx, newRow(receiverA, "biz1")); err != nil || !ins {
		t.Fatalf("first insert (A,biz1) inserted=%v err=%v, want inserted=true", ins, err)
	}
	if ins, err := m.Insert(ctx, newRow(receiverA, "biz1")); err != nil || ins {
		t.Fatalf("dup insert (A,biz1) inserted=%v err=%v, want inserted=false", ins, err)
	}

	// 2. 多接收者：同 type+biz 不同 receiver 独立成行
	if ins, err := m.Insert(ctx, newRow(receiverB, "biz1")); err != nil || !ins {
		t.Fatalf("insert (B,biz1) inserted=%v err=%v, want inserted=true", ins, err)
	}

	// 3. 不同 biz：A 再加一条
	if ins, err := m.Insert(ctx, newRow(receiverA, "biz2")); err != nil || !ins {
		t.Fatalf("insert (A,biz2) inserted=%v err=%v, want inserted=true", ins, err)
	}

	// ListByReceiver(unreadOnly) + CountUnread：A 应有 2 条未读
	listA, err := m.ListByReceiver(ctx, receiverA, true, 0, 50)
	if err != nil {
		t.Fatalf("ListByReceiver A: %v", err)
	}
	if len(listA) != 2 {
		t.Fatalf("A unread list = %d, want 2", len(listA))
	}
	if cnt, _ := m.CountUnread(ctx, receiverA); cnt != 2 {
		t.Fatalf("A unread count = %d, want 2", cnt)
	}
	if cnt, _ := m.CountUnread(ctx, receiverB); cnt != 1 {
		t.Fatalf("B unread count = %d, want 1", cnt)
	}

	// MarkRead 单条：标 A 列表中第一条，未读应降到 1
	if aff, unread, err := m.MarkRead(ctx, receiverA, []uint64{listA[0].Id}); err != nil || aff != 1 || unread != 1 {
		t.Fatalf("MarkRead one affected=%d err=%v, want 1", aff, err)
	}
	if cnt, _ := m.CountUnread(ctx, receiverA); cnt != 1 {
		t.Fatalf("A unread after mark-one = %d, want 1", cnt)
	}
	if ins, err := m.Insert(ctx, newRow(receiverA, "biz3")); err != nil || !ins {
		t.Fatalf("insert biz3: %v", err)
	}
	if aff, unread, err := m.MarkReadByBusiness(ctx, receiverA, []NotificationReadTarget{{NotifyType: marker, BizId: "biz3"}}); err != nil || aff != 1 || unread != 1 {
		t.Fatalf("MarkReadByBusiness affected=%d err=%v", aff, err)
	}
	if cnt, _ := m.CountUnread(ctx, receiverA); cnt != 1 {
		t.Fatalf("business filter cleared unrelated rows: %d", cnt)
	}

	// MarkRead 全部（ids 空）：A 未读清零，且不影响 B
	if _, _, err := m.MarkRead(ctx, receiverA, nil); err != nil {
		t.Fatalf("MarkRead all: %v", err)
	}
	if cnt, _ := m.CountUnread(ctx, receiverA); cnt != 0 {
		t.Fatalf("A unread after mark-all = %d, want 0", cnt)
	}
	if cnt, _ := m.CountUnread(ctx, receiverB); cnt != 1 {
		t.Fatalf("B unread after A mark-all = %d, want 1 (cross-receiver leak)", cnt)
	}
}

func TestNotificationListValidation(t *testing.T) {
	m := &customNotificationModel{}

	_, err := m.ListByReceiver(context.Background(), "01", false, 0, 20)
	if err != ErrInvalidNotificationReceiver {
		t.Fatalf("non-canonical receiver error = %v, want %v", err, ErrInvalidNotificationReceiver)
	}
	for _, pagination := range []struct {
		offset int
		limit  int
	}{{offset: -1, limit: 20}, {offset: 0, limit: -1}, {offset: 0, limit: 101}} {
		_, err = m.ListByReceiver(context.Background(), "1", false, pagination.offset, pagination.limit)
		if err != ErrInvalidNotificationPagination {
			t.Errorf("offset=%d limit=%d error = %v, want %v", pagination.offset, pagination.limit, err, ErrInvalidNotificationPagination)
		}
	}
}
