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
		// ListByNamePage 按昵称模糊分页搜索，返回当前页数据与匹配总数
		ListByNamePage(ctx context.Context, name string, offset, limit int64) ([]*Users, int64, error)
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

// ListByNamePage 按昵称模糊分页搜索（GORM，三库兼容）。
// offset/limit 由调用方计算；返回当前页数据与匹配总数。
func (m *customUsersModel) ListByNamePage(ctx context.Context, name string, offset, limit int64) ([]*Users, int64, error) {
	conn := transaction.GetTransactionOrDB(ctx, db.GetMysqlConn(db.MYSQL_DB_HICHAT2))

	var (
		list  []*Users
		total int64
	)

	tx := conn.WithContext(ctx).Model(&Users{}).Where("nickname LIKE ?", "%"+name+"%")
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return list, 0, nil
	}
	if err := tx.Offset(int(offset)).Limit(int(limit)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
