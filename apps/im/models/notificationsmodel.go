package model

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Notification 公共通知表行模型（运行时读写）。
// 表结构定义见 pkg/db/objects/im.go，迁移走 deploy/sql_init.go。
type Notification struct {
	Id         uint64     `db:"id"`
	ReceiverId string     `db:"receiver_id"`
	NotifyType string     `db:"notify_type"`
	BizId      string     `db:"biz_id"`
	ActorId    string     `db:"actor_id"`
	GroupId    string     `db:"group_id"`
	Title      string     `db:"title"`
	Content    string     `db:"content"`
	Payload    string     `db:"payload"`
	IsRead     int        `db:"is_read"`
	CreatedAt  time.Time  `db:"created_at"`
	ReadAt     *time.Time `db:"read_at"`
}

// TableName GORM 表名
func (Notification) TableName() string {
	return "notifications"
}

type (
	// NotificationModel 公共通知 model
	NotificationModel interface {
		notificationModel
	}

	notificationModel interface {
		// Insert 幂等写入：命中 (receiver_id, notify_type, biz_id) 唯一键则跳过。
		// 返回 inserted=false 表示重复（已存在），用于消费侧判断是否需要推送。
		Insert(ctx context.Context, data *Notification) (inserted bool, err error)
		// ListByReceiver 按接收者分页拉取，unreadOnly 仅未读，按 id 倒序。
		ListByReceiver(ctx context.Context, receiverId string, unreadOnly bool, offset, limit int) ([]*Notification, error)
		// CountUnread 接收者未读数
		CountUnread(ctx context.Context, receiverId string) (int64, error)
		// MarkRead 标记已读；ids 为空表示标记该接收者全部未读为已读。返回受影响行数。
		MarkRead(ctx context.Context, receiverId string, ids []uint64) (int64, error)
	}

	customNotificationModel struct {
		table string
	}
)

// NewNotificationModel 复用进程全局 MySQL 连接（db.GetMysqlConn），无需额外配置接线。
func NewNotificationModel() NotificationModel {
	return &customNotificationModel{
		table: "notifications",
	}
}

func (m *customNotificationModel) conn() *gorm.DB {
	return db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
}

func (m *customNotificationModel) Insert(ctx context.Context, data *Notification) (bool, error) {
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	// ON CONFLICT DO NOTHING：GORM 跨 SQLite/MySQL/PostgreSQL 兼容，命中唯一键不报错也不覆盖。
	res := m.conn().WithContext(ctx).Table(m.table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "receiver_id"}, {Name: "notify_type"}, {Name: "biz_id"}},
			DoNothing: true,
		}).
		Create(data)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (m *customNotificationModel) ListByReceiver(ctx context.Context, receiverId string, unreadOnly bool, offset, limit int) ([]*Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	query := m.conn().WithContext(ctx).Table(m.table).Where("receiver_id = ?", receiverId)
	if unreadOnly {
		query = query.Where("is_read = ?", 0)
	}
	var list []*Notification
	err := query.Order("id desc").Offset(offset).Limit(limit).Find(&list).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return list, nil
}

func (m *customNotificationModel) CountUnread(ctx context.Context, receiverId string) (int64, error) {
	var cnt int64
	err := m.conn().WithContext(ctx).Table(m.table).
		Where("receiver_id = ?", receiverId).Where("is_read = ?", 0).
		Count(&cnt).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	return cnt, nil
}

func (m *customNotificationModel) MarkRead(ctx context.Context, receiverId string, ids []uint64) (int64, error) {
	now := time.Now()
	query := m.conn().WithContext(ctx).Table(m.table).
		Where("receiver_id = ?", receiverId).Where("is_read = ?", 0)
	if len(ids) > 0 {
		query = query.Where("id in ?", ids)
	}
	res := query.Updates(map[string]interface{}{"is_read": 1, "read_at": now})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}
