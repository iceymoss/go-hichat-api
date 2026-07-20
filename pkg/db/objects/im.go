package objects

import (
	"time"
)

// Notification 公共通知表。
// 公共通知通道的落库存储：业务方投递通知事件 -> task/mq 消费 -> 经 im-rpc 落到本表 -> 在线推 ws。
// 幂等键 (receiver_id, notify_type, biz_id) 唯一，对付 Kafka 重投与消费重试。
type Notification struct {
	ID         uint64     `gorm:"primaryKey;column:id;type:BIGINT UNSIGNED;autoIncrement;comment:自增主键"`
	ReceiverID string     `gorm:"column:receiver_id;type:VARCHAR(64);not null;uniqueIndex:uk_recv_type_biz,priority:1;index:idx_recv_read,priority:1;comment:接收者UID"`
	NotifyType string     `gorm:"column:notify_type;type:VARCHAR(64);not null;uniqueIndex:uk_recv_type_biz,priority:2;comment:通知类型 friend.apply/group.accept..."`
	BizID      string     `gorm:"column:biz_id;type:VARCHAR(128);not null;default:'';uniqueIndex:uk_recv_type_biz,priority:3;comment:业务关联ID(幂等键)"`
	ActorID    string     `gorm:"column:actor_id;type:VARCHAR(64);not null;default:'';comment:触发者UID"`
	GroupID    string     `gorm:"column:group_id;type:VARCHAR(64);not null;default:'';comment:群ID(群相关通知)"`
	Title      string     `gorm:"column:title;type:VARCHAR(255);not null;default:'';comment:标题(可空,前端也可用type+i18n渲染)"`
	Content    string     `gorm:"column:content;type:VARCHAR(512);not null;default:'';comment:正文摘要"`
	Payload    string     `gorm:"column:payload;type:TEXT;comment:扩展字段JSON"`
	IsRead     int        `gorm:"column:is_read;type:TINYINT;not null;default:0;index:idx_recv_read,priority:2;comment:0未读 1已读"`
	CreatedAt  *time.Time `gorm:"column:created_at;type:TIMESTAMP;comment:创建时间"`
	ReadAt     *time.Time `gorm:"column:read_at;type:TIMESTAMP;comment:已读时间"`
}

// TableName 指定表名
func (Notification) TableName() string {
	return "notifications"
}

// NotificationReadIntent makes a business notification read durable even when
// the notification event arrives after the mark-read request.
type NotificationReadIntent struct {
	ID         uint64    `gorm:"primaryKey;column:id;autoIncrement"`
	ReceiverID string    `gorm:"column:receiver_id;type:VARCHAR(64);not null;uniqueIndex:uk_notification_read_intent,priority:1"`
	NotifyType string    `gorm:"column:notify_type;type:VARCHAR(64);not null;uniqueIndex:uk_notification_read_intent,priority:2"`
	BizID      string    `gorm:"column:biz_id;type:VARCHAR(128);not null;uniqueIndex:uk_notification_read_intent,priority:3"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
}

func (NotificationReadIntent) TableName() string {
	return "notification_read_intents"
}
