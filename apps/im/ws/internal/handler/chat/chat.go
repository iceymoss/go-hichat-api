package chat

import (
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	libWebsocket "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
)

func Chat(srvCtx *svc.ServiceContext) libWebsocket.HandlerFunc {
	return func(srv *libWebsocket.Server, conn *libWebsocket.Conn, msg *libWebsocket.Message) {
		msg.Data = "pong"
		response := libWebsocket.NewMessageTest(srv, conn, msg.Data)
		response.Method = msg.Method
		err := srv.Send(response, conn)
		srv.Info(err)
	}
}
