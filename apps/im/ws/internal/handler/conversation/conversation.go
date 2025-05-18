package conversation

import (
	"context"
	"github.com/go-viper/mapstructure/v2"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
)

// Chat 单聊处理方法，针对用户处理的消息，
func Chat(srvCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		// 解析mes.data
		var data ws.Chat
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		// 获取当前连接的用户id
		l := logic.NewUserLogic(context.Background(), srv, srvCtx)
		users := srv.GetUsers([]*websocket.Conn{conn})
		if len(users) == 0 {
			zLog.Error("Chat.srv.GetUsers: userConn len is zero")
			return
		}

		if err := l.Chat(&data, users[0]); err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
		}
	}
}
