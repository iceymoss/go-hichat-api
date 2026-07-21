package push

import (
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/go-viper/mapstructure/v2"
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

		fmt.Printf("ws服务端接收到mq推送的数据了:%+v\n", data)

		switch data.ChatType {
		case constants.SingleChatType:
			single(srv, &data, data.RecvId)
		case constants.GroupChatType:
			group(srv, &data)
		}
	}
}

func single(srv *websocket.Server, data *ws.Push, recvId string) error {
	rconn := srv.GetConn([]string{recvId})
	if len(rconn) == 0 {
		// 离线
		return nil
	}
	srv.Infof("push  uid %v", recvId)
	sendMsg := websocket.NewMessage(
		data.SendId, &ws.Chat{
			ConversationId: data.ConversationId,
			ChatType:       data.ChatType,
			RecvId:         recvId,
			SendId:         data.SendId,
			SendTime:       data.SendTime,
			ContentType:    data.ContentType,
			RecalledBy:     data.RecalledBy,
			Msg: ws.Msg{
				MType:       data.MType,
				Content:     data.Content,
				Quote:       data.Quote,
				ReadRecords: data.ReadRecords,
				AtUsers:     data.AtUsers,
				AtAll:       data.AtAll,
			},
		})
	// 把 MongoDB MsgId 写入 WS 帧的顶层 Id 字段，前端据此建立 local_<=>mongoID 的映射
	if data.MsgId != "" {
		sendMsg.Id = data.MsgId
	}
	fmt.Printf("push到客户端的数据: %+v \n", sendMsg)
	return srv.Send(sendMsg, rconn[0])
}

// group 基于并发发送
func group(srv *websocket.Server, data *ws.Push) error {
	for _, id := range data.RecvIdList {
		func(id string) {
			srv.TaskRunner.Schedule(func() {
				single(srv, data, id)
			})
		}(id)
	}
	return nil
}
