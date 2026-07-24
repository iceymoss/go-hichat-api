package main

import (
	"flag"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/middleware"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	pkgConfig "github.com/iceymoss/go-hichat-api/pkg/config"
	httpPkg "github.com/iceymoss/go-hichat-api/pkg/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "apps/social/api/etc/social-sample.yaml", "the config file")

func main() {
	flag.Parse()
	pkgConfig.InitConfig("local", "config")

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()
	httpx.SetErrorHandlerCtx(httpPkg.ErrHandler(c.Name))

	// 使用自定义 CORS 中间件（更完善的跨域支持）
	server.Use(middleware.CORSMiddleware())

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
