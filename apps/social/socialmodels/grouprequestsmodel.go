package socialmodels

import (
	"context"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GroupRequestsModel = (*customGroupRequestsModel)(nil)

type (
	// GroupRequestsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGroupRequestsModel.
	GroupRequestsModel interface {
		groupRequestsModel

		// DeleteByGroupId 解散群时清理入群申请记录（可选）
		DeleteByGroupId(ctx context.Context, groupId string) error
	}

	customGroupRequestsModel struct {
		*defaultGroupRequestsModel
	}
)

// NewGroupRequestsModel returns a model for the database table.
func NewGroupRequestsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) GroupRequestsModel {
	return &customGroupRequestsModel{
		defaultGroupRequestsModel: newGroupRequestsModel(conn, c, opts...),
	}
}

func (m *customGroupRequestsModel) DeleteByGroupId(ctx context.Context, groupId string) error {
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	return mysqlConn.WithContext(ctx).Table(m.table).Where("group_id = ?", groupId).Delete(&GroupRequests{}).Error
}
