package svc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	// Config 服务配置
	Config config.Config

	// 信令服务器
	SignalingServer *handler.SignalingServer

	// 导入各个微服务模块
	Social socialclient.Social
	User   userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	svcCtx := &ServiceContext{
		Config: c,
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}

	// 初始化信令服务器
	svcCtx.SignalingServer = handler.NewSignalingServer(svcCtx)

	return svcCtx
}

func (svcCtx *ServiceContext) GetToken() (string, error) {
	redisConn := db.GetRedisConn()
	res := redisConn.Get(context.Background(), constants.REDIS_SYSTEM_ROOT_TOEKN)
	return res.Val(), res.Err()
}

// Start 启动服务
func (svcCtx *ServiceContext) Start() error {
	// 设置HTTP路由

	
	// 启动HTTP服务器
	go func() {
		fmt.Printf("Starting streaming service on %s\n", svcCtx.Config.ListenOn)
		if err := http.ListenAndServe(svcCtx.Config.ListenOn, nil); err != nil {
			fmt.Printf("Failed to start HTTP server: %v\n", err)
		}

	
	return nil
}

// Stop 停止服务
func (svcCtx *ServiceContext) Stop() error {
	if svcCtx.SignalingServer != nil {
		return svcCtx.SignalingServer.Close()
	}
	return nil
}
