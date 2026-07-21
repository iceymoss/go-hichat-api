package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	SocialRpc zrpc.RpcClientConf //连接rpc服务的，通过服务发现etcd去获取到rpc服务的配置

	UserRpc zrpc.RpcClientConf

	ImRpc zrpc.RpcClientConf

	// MsgRecallTransfer 撤回事件专用的 Kafka topic（生产端）
	MsgRecallTransfer struct {
		Addrs []string
		Topic string
	}

	// Upload 富媒体文件本地存储配置
	Upload struct {
		BasePath string `json:",optional"` // 本地存储路径，如 ./temp
		BaseURL  string `json:",optional"` // 访问URL前缀，如 http://localhost:8887/static
	}
}
