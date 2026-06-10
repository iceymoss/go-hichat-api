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

// RelationNotify 关系变更通知推送：mq 消费者 -> ws -> 客户端。
// 与聊天 push 解耦，单推给 ReceiverId；前端用 method=relation.changed 接收并禁用对应会话输入。
func RelationNotify(svcCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.RelationChanged
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			zLog.Error("RelationNotify.Decode: decode failed", zap.Any("msg", msg), zap.Error(err))
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		rconn := srv.GetConn([]string{data.ReceiverId})
		if len(rconn) == 0 {
			// 接收者离线：上线拉会话/群列表即为失效态，直接返回
			return
		}

		sendMsg := websocket.NewMessage(constants.SYSTEM_ROOT_UID, &data)
		// 显式设置 method，前端用 ws.on('relation.changed') 接收（不污染聊天 push 路由）
		sendMsg.Method = "relation.changed"
		if err := srv.Send(sendMsg, rconn[0]); err != nil {
			zLog.Error("RelationNotify.Send: send failed", zap.String("receiver", data.ReceiverId), zap.Error(err))
		}
	}
}
