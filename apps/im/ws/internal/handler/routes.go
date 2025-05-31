package handler

import (
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler/chat"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler/conversation"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/handler/push"
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
			// 测试ws是否可达
			Method:  "chat.ping",
			Handler: chat.Chat(svc),
		},
		{
			// 聊天处理：客户端到->ws->mq
			Method:  "chat.user",
			Handler: conversation.Chat(svc),
		},
		{
			// push消息: mq->ws->客户端
			Method:  "push",
			Handler: push.Push(svc),
		},
		{
			// 消息阅读处理: 客户端->ws->mq
			Method:  "chat.markChat",
			Handler: conversation.MarkRead(svc),
		},
	})
}
