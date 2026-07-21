package models

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TrendMessageModel = (*customTrendMessageModel)(nil)

type (
	// TrendMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTrendMessageModel.
	TrendMessageModel interface {
		trendMessageModel
		// CreateMsg 写入一条动态消息，返回自增主键
		CreateMsg(ctx context.Context, data *TrendMessage) (uint64, error)
		// ListByReceiver 按接收者分页拉取正常消息（id 倒序，state=1）
		ListByReceiver(ctx context.Context, receiverId uint64, lastId int, limit int) ([]*TrendMessage, error)
		// CountUnreadByType 统计接收者未读消息，按类型分组（state=1 且 is_read=0）
		CountUnreadByType(ctx context.Context, receiverId uint64) (map[uint64]int64, error)
		// MarkAllRead 将接收者全部未读消息置为已读
		MarkAllRead(ctx context.Context, receiverId uint64) error
		// SoftDeleteByTrend 删动态级联：软删该动态下全部消息
		SoftDeleteByTrend(ctx context.Context, trendId uint64) error
		// SoftDeleteByComment 删评论级联：软删该评论相关消息
		SoftDeleteByComment(ctx context.Context, commentId uint64) error
		// SoftDeleteLike 取消点赞级联：软删该用户对该动态的点赞消息
		SoftDeleteLike(ctx context.Context, trendId uint64, actorId uint64) error
	}

	customTrendMessageModel struct {
		*defaultTrendMessageModel
	}
)

// NewTrendMessageModel returns a model for the database table.
func NewTrendMessageModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TrendMessageModel {
	return &customTrendMessageModel{
		defaultTrendMessageModel: newTrendMessageModel(conn, c, opts...),
	}
}

func (m *customTrendMessageModel) CreateMsg(ctx context.Context, data *TrendMessage) (uint64, error) {
	mysqlConn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
	now := time.Now()
	if data.CreateTime.IsZero() {
		data.CreateTime = now
	}
	data.UpdateTime = now
	res := mysqlConn.Table(m.table).Create(data)
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, errors.New("create trend message failed")
	}
	return data.Id, nil
}

func (m *customTrendMessageModel) ListByReceiver(ctx context.Context, receiverId uint64, lastId int, limit int) ([]*TrendMessage, error) {
	if limit <= 0 {
		limit = 20
	}
	mysqlConn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
	query := mysqlConn.Table(m.table).
		Where("receiver_id = ?", receiverId).
		Where("state = ?", constants.TrendMsgStateNormal)
	if lastId > 0 {
		query = query.Where("id < ?", lastId)
	}

	var list []*TrendMessage
	if err := query.Order("id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customTrendMessageModel) CountUnreadByType(ctx context.Context, receiverId uint64) (map[uint64]int64, error) {
	mysqlConn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))

	type row struct {
		Type uint64
		Cnt  int64
	}
	var rows []row
	err := mysqlConn.Table(m.table).
		Select("type, count(*) as cnt").
		Where("receiver_id = ?", receiverId).
		Where("state = ?", constants.TrendMsgStateNormal).
		Where("is_read = ?", constants.TrendMsgUnread).
		Group("type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	ans := make(map[uint64]int64, len(rows))
	for _, r := range rows {
		ans[r.Type] = r.Cnt
	}
	return ans, nil
}

func (m *customTrendMessageModel) MarkAllRead(ctx context.Context, receiverId uint64) error {
	mysqlConn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
	return mysqlConn.Table(m.table).
		Where("receiver_id = ?", receiverId).
		Where("is_read = ?", constants.TrendMsgUnread).
		Where("state = ?", constants.TrendMsgStateNormal).
		Updates(map[string]any{
			"is_read":     constants.TrendMsgRead,
			"update_time": time.Now(),
		}).Error
}

func (m *customTrendMessageModel) SoftDeleteByTrend(ctx context.Context, trendId uint64) error {
	return m.softDelete(ctx, map[string]any{"trend_id": trendId})
}

func (m *customTrendMessageModel) SoftDeleteByComment(ctx context.Context, commentId uint64) error {
	return m.softDelete(ctx, map[string]any{"comment_id": commentId})
}

func (m *customTrendMessageModel) SoftDeleteLike(ctx context.Context, trendId uint64, actorId uint64) error {
	return m.softDelete(ctx, map[string]any{
		"type":     int(constants.TrendMsgLike),
		"trend_id": trendId,
		"actor_id": actorId,
	})
}

// softDelete 将命中 conds 的正常消息置为 state=0（级联软删，幂等）
func (m *customTrendMessageModel) softDelete(ctx context.Context, conds map[string]any) error {
	mysqlConn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
	query := mysqlConn.Table(m.table).Where("state = ?", constants.TrendMsgStateNormal)
	for col, val := range conds {
		query = query.Where(col+" = ?", val)
	}
	return query.Updates(map[string]any{
		"state":       constants.TrendMsgStateDeleted,
		"update_time": time.Now(),
	}).Error
}
