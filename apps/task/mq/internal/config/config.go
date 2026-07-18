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

	// CommonNotifyTransfer 公共通知 topic（落库 + 在线推送，承载好友/群申请等实时通知）
	CommonNotifyTransfer kq.KqConf

	SocialRequestNotification kq.KqConf `json:",optional"`

	// AuthzGate 发送鉴权灰度开关（默认全 false = 现状行为，fail-open）
	AuthzGate struct {
		Enabled    bool
		GroupChat  bool
		SingleChat bool
	}

	Ws struct {
		Host string
	}

	SocialRpc zrpc.RpcClientConf
}
