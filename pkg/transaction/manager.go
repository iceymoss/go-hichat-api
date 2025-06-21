package transaction

import (
	"context"
	"database/sql"
	"github.com/iceymoss/go-hichat-api/pkg/db"

	"github.com/iceymoss/go-hichat-api/pkg/transaction/crdb/crdbgorm"

	"gorm.io/gorm"
)

// Manager 管理数据库事务生命周期和上下文传播
type Manager struct {
	db *gorm.DB
}

// NewManager 创建一个事务管理器实例，自动重试和自动提交或者回滚事务
func NewManager() *Manager {
	return &Manager{
		db: db.GetMysqlConn(db.MYSQL_DB_HICHAT2),
	}
}

// Execute 在事务中执行业务操作
// - ctx: 上下文，用于超时控制和取消操作
// - opts: 事务隔离级别选项
// - operation: 需要在事务中执行业务逻辑的函数
func (m *Manager) Execute(
	ctx context.Context,
	opts *sql.TxOptions,
	operation func(ctx context.Context) error,
) error {
	return crdbgorm.ExecuteTx(ctx, m.db, opts, func(tx *gorm.DB) error {
		// 将事务实例注入上下文
		ctxWithTx := WithTransaction(ctx, tx)
		// 执行业务操作并传递增强后的上下文
		return operation(ctxWithTx)
	})
}
