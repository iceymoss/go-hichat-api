package svc

import (
	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/config"
)

type ServiceContext struct {
	Config config.Config

	// 聊天记录相关
	models.ChatLogModel

	// 会话相关
	models.ConversationsModel

	// 会话下聊天相关
	models.ConversationModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:             c,
		ChatLogModel:       models.NewChatLogModel(),
		ConversationModel:  models.NewConversationModel(),
		ConversationsModel: models.NewConversationsModel(),
	}
}
