package main

import (
	"flag"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "apps/task/mq/etc/mq-local.yaml", "the config file")

// 这是一个消息队列消费者服务
// 实时处理聊天消息（如保存到数据库、推送通知）
func main() {
	flag.Parse()
	var c config.Config

	//加载全局配置
	pkcCfg.InitConfig("local", "", "config")

	conf.MustLoad(*configFile, &c)
	if err := c.SetUp(); err != nil {
		panic(err)
	}

	// 创建服务组（管理多个后台服务）
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	// 初始化服务上下文（传递配置）
	svcCtx := svc.NewServiceContext(c)

	// 创建消息监听器
	listen := handler.NewListen(svcCtx)
	for _, s := range listen.Services() {
		// 将 Kafka 消费者加入服务组
		serviceGroup.Add(s)
	}
	fmt.Println("Starting mqueue server at...")

	defer serviceGroup.Stop()
	serviceGroup.Start()
}
