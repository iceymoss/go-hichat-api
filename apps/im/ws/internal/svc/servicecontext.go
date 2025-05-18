package svc

import (
	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/config"
)

// ServiceContext 服务的上下文和配置
type ServiceContext struct {
	Config config.Config
	model.ChatLogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:       c,
		ChatLogModel: model.NewChatLogModel(),
	}
}
