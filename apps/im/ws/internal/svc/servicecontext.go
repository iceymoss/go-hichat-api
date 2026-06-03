package svc

import (
	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 服务的上下文和配置
type ServiceContext struct {
	Config config.Config

	// 聊天记录
	ChatLogModel model.ChatLogModel

	// 会话
	ConversationsModel model.ConversationsModel

	// mq客户端
	MsgChatTransferClient mq_client.MsgChatTransferClient

	//
	MsgMarkReadTransferClient mq_client.MsgReadTransferClient

	// social：仅用于 @所有人 的生产端角色校验（点查 GetMemberRole）
	Social socialclient.Social
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                    c,
		ChatLogModel:              model.NewChatLogModel(),
		ConversationsModel:        model.NewConversationsModel(),
		MsgChatTransferClient:     mq_client.NewMsgChatTransferClient(c.MsgChatTransfer.Addrs, c.MsgChatTransfer.Topic),
		MsgMarkReadTransferClient: mq_client.NewMsgReadTransferClient(c.MsgMarkRead.Addrs, c.MsgMarkRead.Topic),
		Social:                    socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
	}
}
