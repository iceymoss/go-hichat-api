package push

import (
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/go-viper/mapstructure/v2"
	"go.uber.org/zap"
)

// Notify 公共通知推送：mq 消费者 -> ws -> 客户端。
// 公共通知通道的 ws 下行，单推给 ReceiverId；前端用 method=notify 接收，按 notifyType 分发。
func Notify(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.Notify
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			zLog.Error("Notify.Decode: decode failed", zap.Any("msg", msg), zap.Error(err))
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		rconn := srv.GetConn([]string{data.ReceiverId})
		if len(rconn) == 0 {
			// 接收者离线：通知已落库，上线拉列表可见，直接返回
			return
		}

		sendMsg := websocket.NewMessage(constants.SYSTEM_ROOT_UID, &data)
		// 显式设置 method，前端用 ws.on('notify') 接收
		sendMsg.Method = "notify"
		if err := srv.Send(sendMsg, rconn[0]); err != nil {
			zLog.Error("Notify.Send: send failed", zap.String("receiver", data.ReceiverId), zap.Error(err))
		}
	}
}
