package config

import (
	"github.com/zeromicro/go-queue/kq"
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

	MsgChatTransfer kq.KqConf

	MsgReadTransfer kq.KqConf

	MsgRecallTransfer kq.KqConf

	TrendNotifyTransfer kq.KqConf

	// RelationChangeTransfer 关系变更事件 topic（维护关系缓存 + 推送会话失效）
	RelationChangeTransfer kq.KqConf

	Ws struct {
		Host string
	}

	SocialRpc zrpc.RpcClientConf
}
