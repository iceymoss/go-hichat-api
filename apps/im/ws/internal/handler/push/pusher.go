package push

import (
	"fmt"
	"github.com/go-viper/mapstructure/v2"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

func Push(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.Push

		fmt.Printf("mq消费者将数据推送回来了：%+v\n", msg)

		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			zLog.Error("Push.Decode: decode failed", zap.Any("msg", msg), zap.Error(err))
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		fmt.Printf("data: %+v\n", data)

		rconn := srv.GetConn([]string{data.RecvId})
		if len(rconn) == 0 {
			// 离线
			return
		}

		srv.Infof("push msg %v", msg)

		sendMsg := websocket.NewMessage(
			data.SendId, &ws.Chat{
				ConversationId: data.ConversationId,
				RecvId:         data.RecvId,
				SendId:         data.SendId,
				SendTime:       data.SendTime,

				Msg: ws.Msg{
					MType:   data.MType,
					Content: data.Content,
				},
			})

		fmt.Printf("push到客户端的数据: %+v \n", sendMsg)
		srv.Send(sendMsg, rconn[0])
	}
}
