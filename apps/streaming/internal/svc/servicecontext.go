package svc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
)

type ServiceContext struct {
	// Config 服务配置
	Config config.Config

	// 导入各个微服务模块
	Social socialclient.Social
	User   userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	svcCtx := &ServiceContext{
		Config: c,
		// 暂时注释掉RPC客户端，避免依赖其他服务
		// Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		// User:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}

	return svcCtx
}

func (svcCtx *ServiceContext) GetToken() (string, error) {
	redisConn := db.GetRedisConn()
	res := redisConn.Get(context.Background(), constants.REDIS_SYSTEM_ROOT_TOEKN)
	return res.Val(), res.Err()
}

// Start 启动服务
func (svcCtx *ServiceContext) Start() error {
	// 启动HTTP服务器
	go func() {
		fmt.Printf("Starting streaming service on %s\n", svcCtx.Config.ListenOn)
		if err := http.ListenAndServe(svcCtx.Config.ListenOn, nil); err != nil {
			fmt.Printf("Failed to start HTTP server: %v\n", err)
		}
	}()

	return nil
}

// Stop 停止服务
func (svcCtx *ServiceContext) Stop() error {
	return nil
}
