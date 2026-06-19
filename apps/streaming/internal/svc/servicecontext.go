package svc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	// Config 服务配置
	Config config.Config

	// 跨服务 RPC client：好友/群成员校验、用户资料
	Social socialclient.Social
	User   userclient.User

	// 关系缓存：通话发起的好友/群成员校验（O(1)，miss 回源 RPC）
	RelationCache *relationcache.Cache

	// ImWsClient 以系统 root 身份连入 im ws，推送通话控制信令（push.call -> 客户端 call.signal）
	ImWsClient websocket.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	svcCtx := &ServiceContext{
		Config:        c,
		Social:        socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:          userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		RelationCache: relationcache.New(db.GetRedisConn()),
	}

	// 连接 im ws（同 task/mq 范式：取系统 root token 作鉴权头）
	token, err := svcCtx.GetToken()
	if err != nil {
		fmt.Printf("streaming get system token err: %v\n", err)
	}
	header := http.Header{}
	header.Set("Authorization", token)
	svcCtx.ImWsClient = websocket.NewClient(c.Ws.Host, websocket.WithClientHeader(header))

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
