package svc

import "C"
import (
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext rpc配置，需要调用的模块都需要在这里配置
type ServiceContext struct {
	Config config.Config
	Social socialclient.Social
	User   userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
