package main

import (
	"flag"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/task/cron/internal/svc"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "apps/task/cron/etc/cron-local.yaml", "the config file")

// 这是一个定时任务服务
// 用于执行各种定时任务，如数据清理、统计、通知等
func main() {
	flag.Parse()
	var c config.Config

	// 加载全局配置
	pkcCfg.InitConfig("local", "", "config")

	conf.MustLoad(*configFile, &c)
	if err := c.ApplyDefaultsAndValidate(); err != nil {
		panic(err)
	}
	if err := c.SetUp(); err != nil {
		panic(err)
	}

	// 创建服务组（管理多个后台服务）
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	// 初始化服务上下文（传递配置）
	svcCtx := svc.NewServiceContext(c)

	// 创建定时任务处理器
	cronHandler := handler.NewCronHandler(svcCtx)
	for _, s := range cronHandler.Services() {
		// 将定时任务服务加入服务组
		serviceGroup.Add(s)
	}

	fmt.Println("Starting cron task server...")

	// 启动服务组
	serviceGroup.Start()

	// 阻塞主线程，等待信号
	select {}
}
