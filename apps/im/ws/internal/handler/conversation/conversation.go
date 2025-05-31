package conversation

import (
	"fmt"
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

		fmt.Printf("数据推送给mq: %+v\n", data)

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

// MarkRead 向mq推送已读未读消息的
func MarkRead(svc *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		// todo: 已读未读处理
		var data ws.MarkRead
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}

		err := svc.MsgMarkReadTransferClient.Push(&mq.MsgMarkRead{
			ChatType:       data.ChatType,       // 聊天类型
			ConversationId: data.ConversationId, // 会话id
			SendId:         conn.Uid,            // 消息阅读者
			RecvId:         data.RecvId,         // 接收者
			MsgIds:         data.MsgIds,         // 已被阅读消息id
		})

		if err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
			return
		}
	}
}
