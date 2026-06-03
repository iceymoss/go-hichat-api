package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	SocialRpc zrpc.RpcClientConf //连接rpc服务的，通过服务发现etcd去获取到rpc服务的配置

	// RecallWindowSeconds 发送者本人撤回的时间窗（秒），默认 120；0 表示不限时间（群管理员撤回不受此限制）
	RecallWindowSeconds int64 `json:",default=120"`
}
