package handler

import (
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler/chat"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler/user"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
)

func RegisterHandlers(srv *websocket.Server, svc *svc.ServiceContext) {
	srv.AddRoutes([]websocket.Route{
		{
			Method:  "user.online",
			Handler: user.OnLine(svc),
		},
		{
			Method:  "chat.user",
			Handler: chat.Chat(svc),
		},
	})
}
