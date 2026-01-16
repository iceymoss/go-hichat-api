package socialmodels

import (
	"context"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ FriendRequestsModel = (*customFriendRequestsModel)(nil)

type (
	// FriendRequestsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFriendRequestsModel.
	FriendRequestsModel interface {
		friendRequestsModel
		CountMessageRequests(ctx context.Context, userId string) (int64, error)
	}

	customFriendRequestsModel struct {
		*defaultFriendRequestsModel
	}
)

// NewFriendRequestsModel returns a model for the database table.
func NewFriendRequestsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) FriendRequestsModel {
	return &customFriendRequestsModel{
		defaultFriendRequestsModel: newFriendRequestsModel(conn, c, opts...),
	}
}

// CountMessageRequests 统计需要显示消息气泡的申请数量
// 规则：
// 1. 我发起的申请：handle_result=2（已拒绝）且 status=1（正常显示）的数量
// 2. 我收到的申请：status=1（正常显示）且 handle_result=0（待处理）的数量
func (m *customFriendRequestsModel) CountMessageRequests(ctx context.Context, userId string) (int64, error) {
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)

	var count int64

	// 1. 我发起的申请：handle_result=2（已拒绝）且 status=1（正常显示）
	var sentCount int64
	err := mysqlConn.Table(m.table).
		Where("user_id = ?", userId).
		Where("handle_result = ?", 2).
		Where("status = ?", 1).
		Where("status != ?", 0). // 排除已删除
		Count(&sentCount).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// 2. 我收到的申请：status=1（正常显示）且 handle_result=0（待处理）
	var receivedCount int64
	err = mysqlConn.Table(m.table).
		Where("req_uid = ?", userId).
		Where("status = ?", 1).
		Where("handle_result = ?", 0).
		Where("status != ?", 0). // 排除已删除
		Count(&receivedCount).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	count = sentCount + receivedCount
	return count, nil
}
