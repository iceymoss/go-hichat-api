package svc

import (
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/imclient"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"
	"github.com/iceymoss/go-hichat-api/pkg/storage"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	Social socialclient.Social
	User   userclient.User
	IM     imclient.Im

	// 撤回事件推送生产者（独立 msgRecallTransfer topic）
	MsgRecallTransferClient mq_client.MsgRecallTransferClient

	FileStorage storage.FileStorage
	RPCAuth     *rpcauth.Auth
}

func NewServiceContext(c config.Config) *ServiceContext {
	rpcAuth, err := rpcauth.New(rpcauth.LoadSecret(c.RpcAuthSecret))
	if err != nil {
		panic(fmt.Errorf("im api startup: HICHAT_IM_RPC_AUTH_SECRET is required: %w", err))
	}
	basePath := c.Upload.BasePath
	if basePath == "" {
		basePath = "./temp"
	}
	baseURL := c.Upload.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8887/static"
	}

	return &ServiceContext{
		Config: c,
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		IM: imclient.NewIm(zrpc.MustNewClient(c.ImRpc,
			zrpc.WithUnaryClientInterceptor(rpcAuth.UnaryClientInterceptor()))),
		MsgRecallTransferClient: mq_client.NewMsgRecallTransferClient(c.MsgRecallTransfer.Addrs, c.MsgRecallTransfer.Topic),
		FileStorage:             storage.NewLocalStorage(basePath, baseURL),
		RPCAuth:                 rpcAuth,
	}
}
