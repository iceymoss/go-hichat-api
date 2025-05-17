package chat

import (
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	libWebsocket "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
)

func Chat(srvCtx *svc.ServiceContext) libWebsocket.HandlerFunc {
	return func(srv *libWebsocket.Server, conn *libWebsocket.Conn, msg *libWebsocket.Message) {
		msg.Data = "pong"
		err := srv.SendByUserId(libWebsocket.NewMessage(srv, conn, msg.Data), msg.UserId)
		srv.Info(err)
	}
}
