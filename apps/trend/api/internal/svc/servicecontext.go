package svc

import (
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	mq_client "github.com/iceymoss/go-hichat-api/apps/task/mq/mq_client"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trendservice"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/userclient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Trend  trendservice.TrendService // 不同版本导致
	Social socialclient.Social
	User   userclient.User

	// 动态消息通知推送生产者（独立 trendNotifyTransfer topic）
	TrendNotifyTransferClient mq_client.TrendNotifyTransferClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Trend:  trendservice.NewTrendService(zrpc.MustNewClient(c.TrendRpc)),
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		User:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),

		TrendNotifyTransferClient: mq_client.NewTrendNotifyTransferClient(c.TrendNotifyTransfer.Addrs, c.TrendNotifyTransfer.Topic),
	}
}
