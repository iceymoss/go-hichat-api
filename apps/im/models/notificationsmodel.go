package model

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalidNotificationReceiver   = errors.New("notification receiver must be a canonical positive decimal string")
	ErrInvalidNotificationPagination = errors.New("notification pagination requires offset >= 0 and limit <= 100")
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

type NotificationReadTarget struct {
	NotifyType string
	BizId      string
}

type notificationReadIntent struct {
	Id         uint64    `gorm:"primaryKey;column:id;autoIncrement"`
	ReceiverId string    `gorm:"column:receiver_id"`
	NotifyType string    `gorm:"column:notify_type"`
	BizId      string    `gorm:"column:biz_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (notificationReadIntent) TableName() string {
	return "notification_read_intents"
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
		MarkRead(ctx context.Context, receiverId string, ids []uint64) (affected, unreadCount int64, err error)
		MarkReadByBusiness(ctx context.Context, receiverId string, targets []NotificationReadTarget) (affected, unreadCount int64, err error)
	}

	customNotificationModel struct {
		table  string
		connDB *gorm.DB
	}
)

// NewNotificationModel 复用进程全局 MySQL 连接（db.GetMysqlConn），无需额外配置接线。
func NewNotificationModel() NotificationModel {
	return &customNotificationModel{
		table: "notifications",
	}
}

func (m *customNotificationModel) conn() *gorm.DB {
	if m.connDB != nil {
		return m.connDB
	}
	return db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
}

func (m *customNotificationModel) Insert(ctx context.Context, data *Notification) (bool, error) {
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	var inserted bool
	err := m.conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		intent := &notificationReadIntent{
			ReceiverId: data.ReceiverId,
			NotifyType: data.NotifyType,
			BizId:      data.BizId,
			CreatedAt:  time.Now(),
		}
		lock := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "receiver_id"}, {Name: "notify_type"}, {Name: "biz_id"}},
			DoNothing: true,
		}).Create(intent)
		if lock.Error != nil {
			return lock.Error
		}
		ownsTransientIntent := lock.RowsAffected > 0
		if !ownsTransientIntent {
			data.IsRead = 1
			now := time.Now()
			data.ReadAt = &now
		}

		res := tx.Table(m.table).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "receiver_id"}, {Name: "notify_type"}, {Name: "biz_id"}},
			DoNothing: true,
		}).Create(data)
		if res.Error != nil {
			return res.Error
		}
		inserted = res.RowsAffected > 0
		if ownsTransientIntent {
			return tx.Delete(intent).Error
		}
		return nil
	})
	return inserted, err
}

func (m *customNotificationModel) ListByReceiver(ctx context.Context, receiverId string, unreadOnly bool, offset, limit int) ([]*Notification, error) {
	if !rpcauth.CanonicalUID(receiverId) {
		return nil, ErrInvalidNotificationReceiver
	}
	if offset < 0 || limit < 0 || limit > 100 {
		return nil, ErrInvalidNotificationPagination
	}
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
	if !rpcauth.CanonicalUID(receiverId) {
		return 0, ErrInvalidNotificationReceiver
	}
	var cnt int64
	err := m.conn().WithContext(ctx).Table(m.table).
		Where("receiver_id = ?", receiverId).Where("is_read = ?", 0).
		Count(&cnt).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}
	return cnt, nil
}

func (m *customNotificationModel) MarkRead(ctx context.Context, receiverId string, ids []uint64) (affected, unreadCount int64, err error) {
	if !rpcauth.CanonicalUID(receiverId) {
		return 0, 0, ErrInvalidNotificationReceiver
	}
	err = m.conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Table(m.table).Where("receiver_id = ? AND is_read = ?", receiverId, 0)
		if len(ids) > 0 {
			query = query.Where("id in ?", ids)
		}
		res := query.Updates(map[string]interface{}{"is_read": 1, "read_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		return tx.Table(m.table).Where("receiver_id = ? AND is_read = ?", receiverId, 0).Count(&unreadCount).Error
	})
	return affected, unreadCount, err
}

func (m *customNotificationModel) MarkReadByBusiness(ctx context.Context, receiverId string, targets []NotificationReadTarget) (affected, unreadCount int64, err error) {
	if !rpcauth.CanonicalUID(receiverId) {
		return 0, 0, ErrInvalidNotificationReceiver
	}
	targets = append([]NotificationReadTarget(nil), targets...)
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].NotifyType == targets[j].NotifyType {
			return targets[i].BizId < targets[j].BizId
		}
		return targets[i].NotifyType < targets[j].NotifyType
	})
	err = m.conn().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, target := range targets {
			intent := &notificationReadIntent{ReceiverId: receiverId, NotifyType: target.NotifyType, BizId: target.BizId, CreatedAt: time.Now()}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "receiver_id"}, {Name: "notify_type"}, {Name: "biz_id"}},
				DoNothing: true,
			}).Create(intent).Error; err != nil {
				return err
			}
		}

		pairs := tx.Where("1 = 0")
		for _, target := range targets {
			pairs = pairs.Or("notify_type = ? AND biz_id = ?", target.NotifyType, target.BizId)
		}
		res := tx.Table(m.table).Where("receiver_id = ? AND is_read = ?", receiverId, 0).
			Where(pairs).Updates(map[string]interface{}{"is_read": 1, "read_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		return tx.Table(m.table).Where("receiver_id = ? AND is_read = ?", receiverId, 0).Count(&unreadCount).Error
	})
	return affected, unreadCount, err
}
