package user

import (
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	libWebsocket "github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
)

func OnLine(svc *svc.ServiceContext) libWebsocket.HandlerFunc {
	return func(srv *libWebsocket.Server, conn *libWebsocket.Conn, msg *libWebsocket.Message) {
		uids := srv.GetUsers(nil)
		connList := srv.GetConn(nil)
		err := srv.Send(libWebsocket.NewMessageTest(srv, conn, uids), connList...)
		srv.Info("err ", err)
	}
}
