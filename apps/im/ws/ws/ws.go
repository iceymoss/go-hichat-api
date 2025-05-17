package main

import (
	"flag"
	"fmt"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	websocketServer "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "apps/im/ws/etc/im-local.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := c.SetUp(); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	// 实例化websocket服务
	srv := websocketServer.NewServer(c.ListenOn,
		websocketServer.WithAuthentication(handler.NewJwtAuto(ctx)))
	defer srv.Stop()

	// 处理处理方法
	handler.RegisterHandlers(srv, ctx)

	fmt.Printf("Starting websocket server at %v ...\n", c.ListenOn)
	srv.Start()
}
