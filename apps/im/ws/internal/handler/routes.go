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
			// 动态消息通知: mq->ws->客户端（与聊天 push 解耦）
			Method:  "push.trend",
			Handler: push.TrendNotify(svc),
		},
		{
			// 关系变更通知: mq->ws->客户端（被踢/退群/解散/删好友，前端禁用会话输入）
			Method:  "push.relation",
			Handler: push.RelationNotify(svc),
		},
		{
			// 公共通知: mq->ws->客户端（好友/群申请等，前端 method=notify 按类型分发）
			Method:  "push.notify",
			Handler: push.Notify(svc),
		},
		{
			// 音视频通话控制信令: streaming->ws->客户端（来电/接听/挂断等，前端 method=call.signal）
			Method:  "push.call",
			Handler: push.CallNotify(svc),
		},
		{
			// 消息阅读处理: 客户端->ws->mq
			Method:  "chat.markChat",
			Handler: conversation.MarkRead(svc),
		},
	})
}
