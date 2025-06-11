package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TrendModel = (*customTrendModel)(nil)

type (
	// TrendModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTrendModel.
	TrendModel interface {
		trendModel
	}

	customTrendModel struct {
		*defaultTrendModel
	}
)

// NewTrendModel returns a model for the database table.
func NewTrendModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TrendModel {
	return &customTrendModel{
		defaultTrendModel: newTrendModel(conn, c, opts...),
	}
}
