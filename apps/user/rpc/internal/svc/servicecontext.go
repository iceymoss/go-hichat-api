package svc

import (
	"errors"
	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	UserModels models.UsersModel
	RootToken  string
}

// SetRootToken 从配置或其他来源设置根 Token
func (s *ServiceContext) SetRootToken() error {
	// 示例：从配置读取 Token
	if s.Config.RootToken == "" {
		return errors.New("root token 未配置")
	}
	s.RootToken = s.Config.RootToken
	return nil
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:     c,
		UserModels: models.NewUsersModel(sqlConn, c.Cache),
	}
}
