package user

import (
	"github.com/gorilla/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	libWebsocket "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
)

func OnLine(svc *svc.ServiceContext) libWebsocket.HandlerFunc {
	return func(srv *libWebsocket.Server, conn *websocket.Conn, msg *libWebsocket.Message) {
		uids := srv.GetUsers(nil)
		u := srv.GetUsers([]*websocket.Conn{conn})
		connList := srv.GetConn(nil)
		err := srv.Send(libWebsocket.NewMessage(u[0], uids), connList...)
		srv.Info("err ", err)
	}
}
