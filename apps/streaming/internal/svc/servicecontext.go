package svc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
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

	// 连接 im ws：必须以系统账号 uid=101 登入，绝不能复用 Redis 里某个真实用户的 token，
	// 否则会和该用户在 im ws 上抢同一 uid 槽位互相顶号（addConn 同 uid 顶旧连接）。
	// 用 JWT 密钥现签一个 uid=SYSTEM_ROOT_UID 的 token。
	token, err := ctxdata.GetJwtToken(c.JwtAuth.AccessSecret, time.Now().Unix(), c.JwtAuth.AccessExpire, constants.SYSTEM_ROOT_UID)
	if err != nil {
		fmt.Printf("streaming sign system token err: %v\n", err)
	}
	fmt.Printf("[streaming] im ws auth as system uid=%s tokenLen=%d accessExpire=%d\n",
		constants.SYSTEM_ROOT_UID, len(token), c.JwtAuth.AccessExpire)
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
	// 启动HTTP服务器：绑定失败（如端口被占）直接 panic，避免静默失败导致难排查
	go func() {
		fmt.Printf("Starting streaming service on %s\n", svcCtx.Config.ListenOn)
		if err := http.ListenAndServe(svcCtx.Config.ListenOn, nil); err != nil {
			panic(fmt.Sprintf("streaming HTTP server failed on %s: %v", svcCtx.Config.ListenOn, err))
		}
	}()

	return nil
}

// Stop 停止服务
func (svcCtx *ServiceContext) Stop() error {
	return nil
}
