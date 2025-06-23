package svc

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/user/models"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/iceymoss/go-hichat-api/pkg/db"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	UserModels models.UsersModel
	RootToken  string
}

// SetRootToken 从配置或其他来源设置根 Token
//func (s *ServiceContext) SetRootToken() error {
//	// 示例：从配置读取 Token
//	if s.Config.RootToken == "" {
//		return errors.New("root token 未配置")
//	}
//	s.RootToken = s.Config.RootToken
//	return nil
//}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.Mysql.DataSource)

	return &ServiceContext{
		Config:     c,
		UserModels: models.NewUsersModel(sqlConn, c.Cache),
	}
}

// SetRootToken 设置超级权限的token
func (svcCtx *ServiceContext) SetRootToken() error {
	systemToken, err := ctxdata.GetJwtToken(svcCtx.Config.Jwt.AccessSecret, time.Now().Unix(), 9999999, constants.REDIS_SYSTEM_ROOT_TOEKN)
	if err != nil {
		return err
	}
	redisConn := db.GetRedisConn()
	ctx := context.Background()
	res := redisConn.Set(ctx, constants.REDIS_SYSTEM_ROOT_TOEKN, systemToken, 0)
	return res.Err()
}
