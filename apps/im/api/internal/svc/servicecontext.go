package svc

import (
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/imclient"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/iceymoss/go-hichat-api/pkg/storage"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	Social socialclient.Social
	User   userclient.User
	IM     imclient.Im

	// 撤回事件推送生产者（复用 MsgChatTransfer 通用链路）
	MsgChatTransferClient mq_client.MsgChatTransferClient

	FileStorage storage.FileStorage
}

func NewServiceContext(c config.Config) *ServiceContext {
	basePath := c.Upload.BasePath
	if basePath == "" {
		basePath = "./temp"
	}
	baseURL := c.Upload.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8887/static"
	}

	return &ServiceContext{
		Config:                c,
		Social:                socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:                  userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		IM:                    imclient.NewIm(zrpc.MustNewClient(c.ImRpc)),
		MsgChatTransferClient: mq_client.NewMsgChatTransferClient(c.MsgChatTransfer.Addrs, c.MsgChatTransfer.Topic),
		FileStorage:           storage.NewLocalStorage(basePath, baseURL),
	}
}
