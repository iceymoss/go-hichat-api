package transaction

import (
	"context"

	"gorm.io/gorm"
)

// txContextKey 用于在context中存储事务的标识键
type txContextKey struct{}

// WithTransaction 将事务对象注入到context中，用于事务传递
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// GetTransactionOrDB 从context中获取事务对象，不存在则返回原数据库连接
// 确保始终使用正确的上下文
func GetTransactionOrDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}
