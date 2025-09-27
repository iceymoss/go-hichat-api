package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	service.ServiceConf

	ListenOn string

	Mysql struct {
		DataSource string
	}

	Cache cache.CacheConf

	// 定时任务配置
	Cron struct {
		// 是否启用秒级精度
		WithSeconds bool
		// 任务并发限制
		MaxConcurrency int
		// 任务超时时间(秒)
		TaskTimeout int
	}

	// 各个微服务的RPC配置
	SocialRpc zrpc.RpcClientConf
	UserRpc   zrpc.RpcClientConf
	ImRpc     zrpc.RpcClientConf
	TrendRpc  zrpc.RpcClientConf
}
