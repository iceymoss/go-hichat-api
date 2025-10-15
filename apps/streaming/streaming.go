package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/config"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/handler"
	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "apps/streaming/etc/streaming-local.yaml", "the config file")

// 这是一个流媒体服务
// 用于处理视频和语音通话，支持WebRTC实时通信
func main() {
	flag.Parse()
	var c config.Config

	// 加载全局配置
	pkcCfg.InitConfig("local", "", "config")

	conf.MustLoad(*configFile, &c)
	if err := c.SetUp(); err != nil {
		panic(err)
	}

	fmt.Printf("✅ ListenOn: %s\n", c.ListenOn)
	fmt.Printf("WebRTC.IceServers 数量: %d\n", len(c.WebRTC.IceServers))

	// 初始化服务上下文
	svcCtx := svc.NewServiceContext(c)

	// 初始化信令服务器
	signalingServer := handler.NewSignalingServer(svcCtx)

	// 设置HTTP路由
	http.HandleFunc("/ws", signalingServer.HandleWebSocket)

	// 启动服务
	if err := svcCtx.Start(); err != nil {
		panic(err)
	}

	fmt.Println("Streaming service started successfully!")
	fmt.Printf("WebSocket endpoint: ws://%s/ws\n", c.ListenOn)
	fmt.Println("Press Ctrl+C to stop the service")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down streaming service...")

	// 停止服务
	if err := signalingServer.Close(); err != nil {
		fmt.Printf("Error stopping signaling server: %v\n", err)
	}

	if err := svcCtx.Stop(); err != nil {
		fmt.Printf("Error stopping service: %v\n", err)
	}

	fmt.Println("Streaming service stopped")
}
