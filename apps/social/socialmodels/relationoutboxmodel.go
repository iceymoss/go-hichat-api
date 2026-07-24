package socialmodels

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"gorm.io/gorm"
)

// RelationOutboxModel 关系变更发件箱的运行时读写。
// InsertTx 必须在调用方的同一事务里写入（与关系变更原子提交）；ListPending/MarkSent 供 relay 轮询投递。
type RelationOutboxModel interface {
	// InsertTx 在调用方事务 tx 内插入一条待投递事件，保证与关系变更原子提交。
	InsertTx(tx *gorm.DB, row *objects.RelationOutbox) error
	// ListPending 取最早的若干条待投递事件，按 id 升序（id 即版本，保证投递有序）。
	ListPending(ctx context.Context, limit int) ([]*objects.RelationOutbox, error)
	// MarkSent 批量标记事件已投递。
	MarkSent(ctx context.Context, ids []uint64) error
}

type defaultRelationOutboxModel struct {
	conn  *gorm.DB
	table string
}

// NewRelationOutboxModel 用 hichat2 的 GORM 连接构造发件箱模型。
func NewRelationOutboxModel() RelationOutboxModel {
	return NewRelationOutboxModelWithDB(db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
}

func NewRelationOutboxModelWithDB(conn *gorm.DB) RelationOutboxModel {
	return &defaultRelationOutboxModel{
		conn:  conn,
		table: objects.RelationOutbox{}.TableName(),
	}
}

func (m *defaultRelationOutboxModel) InsertTx(tx *gorm.DB, row *objects.RelationOutbox) error {
	return tx.Table(m.table).Create(row).Error
}

func (m *defaultRelationOutboxModel) ListPending(ctx context.Context, limit int) ([]*objects.RelationOutbox, error) {
	var rows []*objects.RelationOutbox
	err := m.conn.WithContext(ctx).Table(m.table).
		Where("status = ?", 0).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *defaultRelationOutboxModel) MarkSent(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return m.conn.WithContext(ctx).Table(m.table).
		Where("id IN ?", ids).
		Updates(map[string]any{"status": 1, "sent_at": now}).Error
}
