package conversation

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/wuid"
	"time"
)

// Chat 单聊处理方法，针对用户处理的消息，
func Chat(srvCtx *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		// 解析mes.data
		var data ws.Chat
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		if data.ConversationId == "" {
			switch data.ChatType {
			case constants.SingleChatType: // 单聊
				//获取用户id
				users := srv.GetUsers([]*websocket.Conn{conn})
				if len(users) == 0 {
					zLog.Error("Chat.srv.GetUsers: userConn len is zero")
					return
				}
				// 生成会话id
				data.ConversationId = wuid.CombineId(data.RecvId, users[0])
			case constants.GroupChatType: // 群聊
				//todo: 群聊待处理
				data.ConversationId = data.RecvId
			}
		}

		// 向mq推送数据
		err := srvCtx.MsgChatTransferClient.Push(&mq.MsgChatTransfer{
			ChatType:       data.ChatType,
			ConversationId: data.ConversationId,
			SendId:         conn.Uid,
			RecvId:         data.RecvId,
			MsgType:        data.MType,
			MsgContent:     data.Content,
			SendTime:       time.Now().UnixNano(),
		})
		if err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
		}
	}
}
