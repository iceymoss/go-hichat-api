package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	libConfig "github.com/iceymoss/go-hichat-api/pkg/config"
	httpPkg "github.com/iceymoss/go-hichat-api/pkg/http"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "apps/user/api/etc/user-sample.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 配置静态文件服务（需要在创建服务器之前）
	// 从配置文件路径推断项目根目录
	configPath := ""
	if *configFile != "" {
		absConfigFile, err := filepath.Abs(*configFile)
		if err == nil {
			configDir := filepath.Dir(absConfigFile) // apps/user/api/etc
			configDir = filepath.Dir(configDir)      // apps/user/api
			configDir = filepath.Dir(configDir)      // apps/user
			configDir = filepath.Dir(configDir)      // apps
			configDir = filepath.Dir(configDir)      // 项目根目录
			configPath = filepath.Join(configDir, "config")
		} else {
			configPath = "./config"
		}
	} else {
		configPath = "./config"
	}
	libConfig.InitConfig("local", "api", configPath)
	uploadConfig := libConfig.ServiceConf.Upload
	staticPath := uploadConfig.BasePath
	if staticPath == "" {
		staticPath = "./temp"
	}

	// 确保使用绝对路径
	absStaticPath, err := filepath.Abs(staticPath)
	if err != nil {
		absStaticPath = staticPath
	}
	fmt.Printf("Static files absolute path: %s\n", absStaticPath)

	// 使用 go-zero 的 WithFileServer 选项配置静态文件服务
	staticDir := http.Dir(absStaticPath)
	server := rest.MustNewServer(c.RestConf,
		rest.WithCors("*"),
		rest.WithFileServer("/static", staticDir),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	//server.Use(middleware.CORSMiddleware())

	handler.RegisterHandlers(server, ctx)

	httpx.SetErrorHandlerCtx(httpPkg.ErrHandler(c.Name))
	httpx.SetOkHandler(httpPkg.OkHandler)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	fmt.Printf("Static files served from: %s\n", staticPath)
	server.Start()
}
