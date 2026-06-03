package conversation

import (
	"context"
	"fmt"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/wuid"

	"github.com/go-viper/mapstructure/v2"
	"github.com/pkg/errors"
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

		// @ 仅在群聊生效；@所有人在生产端做角色校验，非管理员降级（fail-closed，不阻断消息本体）
		atUsers, atAll := resolveMentions(srvCtx, conn.Uid, &data)

		fmt.Printf("数据推送给mq: %+v\n", data)

		// 向mq推送数据
		err := srvCtx.MsgChatTransferClient.Push(&mq.MsgChatTransfer{
			ChatType:       data.ChatType,
			ConversationId: data.ConversationId,
			SendId:         conn.Uid,
			RecvId:         data.RecvId,
			MsgType:        data.MType,
			MsgContent:     data.Content,
			Quote:          data.Quote,
			SendTime:       time.Now().UnixNano(),
			AtUsers:        atUsers,
			AtAll:          atAll,
		})
		if err != nil {
			srv.Send(websocket.NewErrMessage(err), conn)
		}
	}
}

// resolveMentions 计算最终生效的 @ 信息：
//   - 非群聊：清空 @，不生效
//   - @所有人：仅当发送者是群管理员/群主时保留；否则降级丢弃（前端为主门禁，这里是防御）
//   - GetMemberRole 调用带短超时，失败时 fail-closed（按非管理员处理），绝不阻断消息本体
func resolveMentions(srvCtx *svc.ServiceContext, senderUid string, data *ws.Chat) ([]string, bool) {
	if data.ChatType != constants.GroupChatType {
		return nil, false
	}
	atUsers := data.AtUsers
	atAll := data.AtAll
	if !atAll {
		return atUsers, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	role, err := srvCtx.Social.GetMemberRole(ctx, &socialclient.GetMemberRoleReq{
		GroupId: data.RecvId, // 群聊 RecvId 即 groupId
		UserId:  senderUid,
	})
	if err != nil {
		zLog.Error(fmt.Sprintf("resolveMentions: get member role failed, drop atAll. groupId=%s uid=%s err=%v", data.RecvId, senderUid, err))
		return atUsers, false
	}
	if role.IsMember && role.RoleLevel >= int32(constants.ManagerGroupRoleLevel) {
		return atUsers, true
	}
	return atUsers, false
}

// MarkRead 向mq推送已读未读消息的
func MarkRead(svc *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
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
			ReadRecords:    data.ReadRecords,    // 消息阅读状态 消息id => 当前用户阅读状态
		})

		if err != nil {
			errStr := errors.New(fmt.Sprintf("push mq fialed: %s", err.Error()))
			srv.Send(websocket.NewErrMessage(errStr), conn)
			return
		}
	}
}
