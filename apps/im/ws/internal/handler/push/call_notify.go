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

// CallNotify 音视频通话控制信令推送：streaming 服务（以系统 root 身份连入）-> im ws -> 客户端。
// 与聊天 push 解耦，单推给 ReceiverId；前端用 method=call.signal 接收，按 event 分发来电/接听/挂断等。
// 瞬时态信令，不落库（与 push.notify 区别）；接收者离线即丢，由超时转“未接来电”消息补偿。
func CallNotify(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.CallSignal
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			zLog.Error("CallNotify.Decode: decode failed", zap.Any("msg", msg), zap.Error(err))
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		rconn := srv.GetConn([]string{data.ReceiverId})
		if len(rconn) == 0 {
			// 接收者离线：瞬时信令直接丢弃（来电场景由 streaming 侧超时落“未接来电”消息补偿）
			zLog.Info("CallNotify: receiver offline, drop",
				zap.String("receiver", data.ReceiverId), zap.String("event", data.Event), zap.String("callId", data.CallId))
			return
		}

		sendMsg := websocket.NewMessage(constants.SYSTEM_ROOT_UID, &data)
		// 显式设置 method，前端用 ws.on('call.signal') 接收（不污染聊天 push 路由）
		sendMsg.Method = "call.signal"
		if err := srv.Send(sendMsg, rconn[0]); err != nil {
			zLog.Error("CallNotify.Send: send failed", zap.String("receiver", data.ReceiverId), zap.String("event", data.Event), zap.Error(err))
			return
		}
		zLog.Info("CallNotify: forwarded",
			zap.String("receiver", data.ReceiverId), zap.String("event", data.Event), zap.String("callId", data.CallId))
	}
}
