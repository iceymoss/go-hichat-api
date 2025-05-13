package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	SocialRpc zrpc.RpcClientConf //连接rpc服务的，通过服务发现etcd去获取到rpc服务的配置

	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
}
