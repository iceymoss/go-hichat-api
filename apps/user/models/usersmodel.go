package models

import (
	"context"
	"database/sql"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UsersModel = (*customUsersModel)(nil)

type (
	// UsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersModel.
	UsersModel interface {
		usersModel
	}

	customUsersModel struct {
		*defaultUsersModel
	}
)

// NewUsersModel returns a model for the database table.
func NewUsersModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UsersModel {
	return &customUsersModel{
		defaultUsersModel: newUsersModel(conn, c, opts...),
	}
}

// Create 重写 Create 方法，处理 Email 为 NULL 的情况
func (m *customUsersModel) Create(ctx context.Context, data *Users) error {
	mysqlConn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))
	
	// 如果 Email 无效（NULL），使用 Omit 来忽略它，让数据库使用 NULL
	if !data.Email.Valid || data.Email.String == "" {
		ret := mysqlConn.Model(&Users{}).Omit("email").Create(&data)
		if ret.Error != nil {
			return ret.Error
		}
		if ret.RowsAffected == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	
	// Email 有效，使用默认的 Create 方法
	return m.defaultUsersModel.Create(ctx, data)
}
