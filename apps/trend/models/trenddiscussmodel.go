package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TrendDiscussModel = (*customTrendDiscussModel)(nil)

type (
	// TrendDiscussModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTrendDiscussModel.
	TrendDiscussModel interface {
		trendDiscussModel
	}

	customTrendDiscussModel struct {
		*defaultTrendDiscussModel
	}
)

// NewTrendDiscussModel returns a model for the database table.
func NewTrendDiscussModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TrendDiscussModel {
	return &customTrendDiscussModel{
		defaultTrendDiscussModel: newTrendDiscussModel(conn, c, opts...),
	}
}
