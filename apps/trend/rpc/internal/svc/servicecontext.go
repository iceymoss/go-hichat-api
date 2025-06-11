package svc

import (
	"github.com/iceymoss/go-hichat-api/apps/trend/models"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	Trend models.TrendModel

	TrendDiscuss models.TrendDiscussModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:       c,
		Trend:        models.NewTrendModel(sqlConn, c.Cache),
		TrendDiscuss: models.NewTrendDiscussModel(sqlConn, c.Cache),
	}
}
