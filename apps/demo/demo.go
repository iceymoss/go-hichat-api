package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/iceymoss/go-hichat-api/apps/demo/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/demo/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/demo/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/demo.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 注册 API 路由（WebSocket 等）
	handler.RegisterHandlers(server, ctx)

	// 静态文件服务
	staticDir := "./static"

	// app.js
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/app.js",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			http.ServeFile(w, r, filepath.Join(staticDir, "app.js"))
		}),
	})

	// 首页 - 必须放在最后，避免覆盖其他路由
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
		}),
	})

	fmt.Printf("🎥 视频通话服务启动成功!\n")
	fmt.Printf("📡 信令服务器: ws://%s:%d/ws\n", c.Host, c.Port)
	fmt.Printf("🌐 访问地址: http://%s:%d\n", c.Host, c.Port)
	fmt.Printf("🔧 状态接口: http://%s:%d/status\n", c.Host, c.Port)

	server.Start()
}
