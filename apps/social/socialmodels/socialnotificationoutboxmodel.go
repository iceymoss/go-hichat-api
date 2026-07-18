package socialmodels

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"gorm.io/gorm"
)

type SocialNotificationOutboxModel interface {
	InsertTx(*gorm.DB, *objects.SocialNotificationOutbox) error
	ListDue(context.Context, time.Time, int) ([]*objects.SocialNotificationOutbox, error)
	MarkSent(context.Context, uint64, time.Time) error
	MarkFailed(context.Context, uint64, int, *time.Time, string, bool) error
	CountByStatus(context.Context, int) (int64, error)
	ReplayDead(context.Context, []uint64) (int64, error)
}

func (m *defaultSocialNotificationOutboxModel) CountByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	err := m.conn.WithContext(ctx).Model(&objects.SocialNotificationOutbox{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (m *defaultSocialNotificationOutboxModel) ReplayDead(ctx context.Context, ids []uint64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := m.conn.WithContext(ctx).Model(&objects.SocialNotificationOutbox{}).Where("id IN ? AND status = ?", ids, 2).
		Updates(map[string]any{"status": 0, "attempts": 0, "next_retry_at": nil, "last_error": "", "sent_at": nil})
	return result.RowsAffected, result.Error
}

type defaultSocialNotificationOutboxModel struct{ conn *gorm.DB }

func NewSocialNotificationOutboxModel() SocialNotificationOutboxModel {
	return &defaultSocialNotificationOutboxModel{conn: db.GetMysqlConn(db.MYSQL_DB_HICHAT2)}
}

func (m *defaultSocialNotificationOutboxModel) InsertTx(tx *gorm.DB, row *objects.SocialNotificationOutbox) error {
	return tx.Create(row).Error
}

func (m *defaultSocialNotificationOutboxModel) ListDue(ctx context.Context, now time.Time, limit int) ([]*objects.SocialNotificationOutbox, error) {
	var rows []*objects.SocialNotificationOutbox
	err := m.conn.WithContext(ctx).Where("status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)", 0, now).Order("id ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (m *defaultSocialNotificationOutboxModel) MarkSent(ctx context.Context, id uint64, now time.Time) error {
	return m.conn.WithContext(ctx).Model(&objects.SocialNotificationOutbox{}).Where("id = ? AND status = ?", id, 0).
		Updates(map[string]any{"status": 1, "sent_at": now, "last_error": ""}).Error
}

func (m *defaultSocialNotificationOutboxModel) MarkFailed(ctx context.Context, id uint64, attempts int, next *time.Time, message string, dead bool) error {
	if len(message) > 512 {
		message = message[:512]
	}
	status := 0
	if dead {
		status = 2
	}
	return m.conn.WithContext(ctx).Model(&objects.SocialNotificationOutbox{}).Where("id = ? AND status = ?", id, 0).
		Updates(map[string]any{"status": status, "attempts": attempts, "next_retry_at": next, "last_error": message}).Error
}
