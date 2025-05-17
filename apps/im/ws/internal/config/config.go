package config

import (
	"github.com/zeromicro/go-zero/core/service"
)

type Config struct {

	// 使用go-zero中提供的功能完成对服务的开发及监听工作
	service.ServiceConf

	// 服务监听地址
	ListenOn string
}
