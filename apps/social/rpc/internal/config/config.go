package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	Cache cache.CacheConf

	ImRpc zrpc.RpcClientConf

	UserRpc zrpc.RpcClientConf

	// RelationChangeTransfer 关系变更事件 Kafka 生产端（relay 投递用）
	RelationChangeTransfer struct {
		Addrs []string
		Topic string
	}

	// CommonNotifyTransfer 公共通知事件 Kafka 生产端（好友/群申请等实时通知）
	CommonNotifyTransfer struct {
		Addrs []string
		Topic string
	}
}
