package push

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
)

func Push(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.Push
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		rconn := srv.GetConn([]string{data.RecvId})
		if len(rconn) == 0 {
			// 离线
			return
		}

		srv.Infof("push msg %v", msg)

		sendMsg := websocket.NewMessage(
			data.SendId, &ws.Chat{
				ConversationId: data.ConversationId,
				Msg: ws.Msg{
					MType:   data.MType,
					Content: data.Content,
				},
			})

		srv.Send(sendMsg, rconn[0])
	}
}
