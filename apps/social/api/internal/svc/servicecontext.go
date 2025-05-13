package svc

import "C"
import (
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Social socialclient.Social
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
	}
}
