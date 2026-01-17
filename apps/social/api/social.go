package main

import (
	"flag"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/middleware"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "apps/social/api/etc/social-sample.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	// 使用自定义 CORS 中间件（更完善的跨域支持）
	server.Use(middleware.CORSMiddleware())

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
