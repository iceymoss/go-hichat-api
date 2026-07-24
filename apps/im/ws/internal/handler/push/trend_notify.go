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

// TrendNotify 动态消息通知推送：mq 消费者 -> ws -> 客户端。
// 与聊天 push 解耦，单推给 ReceiverId；前端用 method=trend.notify 接收。
func TrendNotify(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.TrendNotify
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			zLog.Error("TrendNotify.Decode: decode failed", zap.Any("msg", msg), zap.Error(err))
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		rconn := srv.GetConn([]string{data.ReceiverId})
		if len(rconn) == 0 {
			// 接收者离线：消息已落库，上线拉列表可见，直接返回
			return
		}

		sendMsg := websocket.NewMessage(constants.SYSTEM_ROOT_UID, &data)
		// 显式设置 method，前端用 ws.on('trend.notify') 接收（不污染聊天的 push 路由）
		sendMsg.Method = "trend.notify"
		if err := srv.Send(sendMsg, rconn...); err != nil {
			zLog.Error("TrendNotify.Send: send failed", zap.String("receiver", data.ReceiverId), zap.Error(err))
		}
	}
}
