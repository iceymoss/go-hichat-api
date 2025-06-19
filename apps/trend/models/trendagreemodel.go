package models

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TrendAgreeModel = (*customTrendAgreeModel)(nil)

type (
	// TrendAgreeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTrendAgreeModel.
	TrendAgreeModel interface {
		trendAgreeModel
	}

	customTrendAgreeModel struct {
		*defaultTrendAgreeModel
	}
)

// NewTrendAgreeModel returns a model for the database table.
func NewTrendAgreeModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TrendAgreeModel {
	return &customTrendAgreeModel{
		defaultTrendAgreeModel: newTrendAgreeModel(conn, c, opts...),
	}
}
