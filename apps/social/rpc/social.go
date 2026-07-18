package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/relay"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/server"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/interceptor/rpcserver"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "apps/social/rpc/etc/social-sample.yaml", "the config file")

func main() {
	flag.Parse()

	pkcCfg.InitConfig("local", "", "config")

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	fmt.Println("config:", pkcCfg.ServiceConf)

	// 启动关系变更发件箱投递器（后台 goroutine，进程退出时随 ctx 取消优雅停止）
	relayCtx, cancelRelay := context.WithCancel(context.Background())
	defer cancelRelay()
	go relay.New(ctx).Start(relayCtx)
	go relay.NewNotification(ctx).Start(relayCtx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		social.RegisterSocialServer(grpcServer, server.NewSocialServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(rpcserver.LogInterceptor)
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
