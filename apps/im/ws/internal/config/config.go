package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"time"
)

type Config struct {

	// 使用go-zero中提供的功能完成对服务的开发及监听工作
	service.ServiceConf

	// 服务监听地址
	ListenOn string

	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	MsgChatTransfer kafkaCfg

	MsgMarkRead kafkaCfg

	// SocialRpc 仅用于 @所有人 的生产端角色校验（点查 GetMemberRole）
	SocialRpc zrpc.RpcClientConf

	// AuthzGate 发送鉴权灰度开关（默认全 false = 现状行为，fail-open）
	AuthzGate struct {
		Enabled    bool
		GroupChat  bool
		SingleChat bool
	}
	Presence struct {
		NodeId           string        `json:",optional"`
		TTL              time.Duration `json:",default=5m"`
		Refresh          time.Duration `json:",default=2m"`
		HeartbeatTimeout time.Duration `json:",default=75s"`
	}
}

type kafkaCfg struct {
	Addrs []string
	Topic string
}
