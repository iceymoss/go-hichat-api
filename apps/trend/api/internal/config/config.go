package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	TrendRpc zrpc.RpcClientConf // 动态rpc

	SocialRpc zrpc.RpcClientConf //连接rpc服务的，通过服务发现etcd去获取到rpc服务的配置

	UserRpc zrpc.RpcClientConf

	ImRpc zrpc.RpcClientConf

	// TrendNotifyTransfer 动态消息通知专用 Kafka topic（生产端）
	TrendNotifyTransfer struct {
		Addrs []string
		Topic string
	}

	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
}
