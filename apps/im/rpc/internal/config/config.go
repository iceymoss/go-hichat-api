package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	SocialRpc zrpc.RpcClientConf //连接rpc服务的，通过服务发现etcd去获取到rpc服务的配置

}
