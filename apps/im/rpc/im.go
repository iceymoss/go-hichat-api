package main

import (
	"flag"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/server"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "apps/im/rpc/etc/im-sample.yaml", "the config file")

func main() {
	flag.Parse()

	pkcCfg.InitConfig("local", "", "config")

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		im.RegisterImServer(grpcServer, server.NewImServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(ctx.RPCAuth.UnaryServerInterceptor(ctx.RPCReplay,
		im.Im_CreateNotification_FullMethodName,
		im.Im_ListNotifications_FullMethodName,
		im.Im_MarkNotificationsRead_FullMethodName,
		im.Im_GetUnreadNotificationCount_FullMethodName,
	))
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
