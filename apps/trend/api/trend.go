package main

import (
	"flag"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/trend-local.yaml", "the config file")

func main() {
	flag.Parse()

	pkcCfg.InitConfig("local", "", "config")

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
