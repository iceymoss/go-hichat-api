package model

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSQLiteNotificationModel(t *testing.T) (*customNotificationModel, *gorm.DB) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "notifications.sqlite") + "?_busy_timeout=5000"
	conn, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := conn.DB()
	require.NoError(t, err)
	// SQLite permits only one writer. Calls remain concurrent at the model API
	// boundary while database/sql serializes their transactions on this DB.
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, conn.AutoMigrate(&objects.Notification{}, &objects.NotificationReadIntent{}))
	return &customNotificationModel{table: "notifications", connDB: conn}, conn
}

func TestNotificationReadIntentOrderingAndIsolation(t *testing.T) {
	ctx := context.Background()
	target := NotificationReadTarget{NotifyType: "friend.apply", BizId: "friend:9007199254740993:apply"}

	t.Run("mark before insert", func(t *testing.T) {
		m, _ := newSQLiteNotificationModel(t)
		affected, unread, err := m.MarkReadByBusiness(ctx, "1", []NotificationReadTarget{target})
		require.NoError(t, err)
		require.Zero(t, affected)
		require.Zero(t, unread)

		row := &Notification{ReceiverId: "1", NotifyType: target.NotifyType, BizId: target.BizId}
		inserted, err := m.Insert(ctx, row)
		require.NoError(t, err)
		require.True(t, inserted)
		require.Equal(t, 1, row.IsRead)
		require.NotNil(t, row.ReadAt)
	})

	t.Run("insert before mark and idempotency", func(t *testing.T) {
		m, _ := newSQLiteNotificationModel(t)
		row := &Notification{ReceiverId: "1", NotifyType: target.NotifyType, BizId: target.BizId}
		inserted, err := m.Insert(ctx, row)
		require.NoError(t, err)
		require.True(t, inserted)

		affected, unread, err := m.MarkReadByBusiness(ctx, "1", []NotificationReadTarget{target, target})
		require.NoError(t, err)
		require.EqualValues(t, 1, affected)
		require.Zero(t, unread)
		affected, unread, err = m.MarkReadByBusiness(ctx, "1", []NotificationReadTarget{target})
		require.NoError(t, err)
		require.Zero(t, affected)
		require.Zero(t, unread)
	})

	t.Run("exact pairs and receiver isolation", func(t *testing.T) {
		m, _ := newSQLiteNotificationModel(t)
		targets := []NotificationReadTarget{
			{NotifyType: "friend.apply", BizId: "friend:1:apply"},
			{NotifyType: "group.accept", BizId: "group:2:accept"},
		}
		for _, receiver := range []string{"1", "2"} {
			for _, row := range []*Notification{
				{ReceiverId: receiver, NotifyType: "friend.apply", BizId: "friend:1:apply"},
				{ReceiverId: receiver, NotifyType: "group.accept", BizId: "group:2:accept"},
				{ReceiverId: receiver, NotifyType: "friend.apply", BizId: "group:2:accept"},
				{ReceiverId: receiver, NotifyType: "group.accept", BizId: "friend:1:apply"},
			} {
				inserted, err := m.Insert(ctx, row)
				require.NoError(t, err)
				require.True(t, inserted)
			}
		}

		affected, unread, err := m.MarkReadByBusiness(ctx, "1", targets)
		require.NoError(t, err)
		require.EqualValues(t, 2, affected)
		require.EqualValues(t, 2, unread)
		otherUnread, err := m.CountUnread(ctx, "2")
		require.NoError(t, err)
		require.EqualValues(t, 4, otherUnread)
	})
}

func TestNotificationReadIntentConcurrentConvergence(t *testing.T) {
	for iteration := 0; iteration < 5; iteration++ {
		t.Run(strconv.Itoa(iteration), func(t *testing.T) {
			m, conn := newSQLiteNotificationModel(t)
			ctx := context.Background()
			target := NotificationReadTarget{NotifyType: "group.reject", BizId: "group:42:reject"}
			start := make(chan struct{})
			errs := make(chan error, 2)
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, err := m.Insert(ctx, &Notification{ReceiverId: "7", NotifyType: target.NotifyType, BizId: target.BizId})
				errs <- err
			}()
			go func() {
				defer wg.Done()
				<-start
				_, _, err := m.MarkReadByBusiness(ctx, "7", []NotificationReadTarget{target})
				errs <- err
			}()
			close(start)
			wg.Wait()
			close(errs)
			for err := range errs {
				require.NoError(t, err)
			}

			var row Notification
			require.NoError(t, conn.Where("receiver_id = ?", "7").First(&row).Error)
			require.Equal(t, 1, row.IsRead)
			require.NotNil(t, row.ReadAt)
			count, err := m.CountUnread(ctx, "7")
			require.NoError(t, err)
			require.Zero(t, count)
		})
	}
}
