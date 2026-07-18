package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	websocketServer "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/presence"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "apps/im/ws/etc/im-local.yaml", "the config file")

func main() {
	flag.Parse()

	//加载全局配置
	pkcCfg.InitConfig("local", "", "config")

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := c.SetUp(); err != nil {
		panic(err)
	}

	ctx := svc.NewServiceContext(c)

	option := websocketServer.WithAuthentication(handler.NewJwtAuto(ctx))
	nodeID := c.Presence.NodeId
	if nodeID == "" {
		host, _ := os.Hostname()
		nodeID = host + "@" + c.ListenOn
	}
	presenceOption := websocketServer.WithPresence(nodeID, presence.New(db.GetRedisConn()), c.Presence.TTL, c.Presence.Refresh)
	heartbeatOption := websocketServer.WithMaxConnectionIdle(c.Presence.HeartbeatTimeout)
	//option = websocketServer.WithAck(websocketServer.RigorAck)

	// 实例化websocket服务
	srv := websocketServer.NewServer(c.ListenOn, option, presenceOption, heartbeatOption)
	defer srv.Stop()

	// 处理处理方法
	handler.RegisterHandlers(srv, ctx)

	fmt.Printf("Starting websocket server at %v ...\n", c.ListenOn)
	srv.Start()
}
