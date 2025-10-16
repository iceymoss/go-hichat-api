package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	// STUN 服务器配置
	StunServers []string
}
