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
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"
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

		// 单聊发送鉴权闸门（灰度开关 + fail-open）：删好友后不得再发消息。命中拦截则不投递。
		if data.ChatType == constants.SingleChatType && !singleSendAllowed(srvCtx, conn.Uid, data.RecvId) {
			zLog.Error("single send blocked: not friend")
			srv.Send(websocket.NewErrMessageWithId(msg.Id, errors.New("对方不是你的好友，无法发送消息")), conn)
			return
		}

		// 群发送鉴权闸门（灰度开关 + fail-open）：被移出群后不得再发消息。命中拦截则不投递。
		if data.ChatType == constants.GroupChatType && !groupSendAllowed(srvCtx, conn.Uid, data.RecvId) {
			zLog.Error("group send blocked: not member")
			srv.Send(websocket.NewErrMessageWithId(msg.Id, errors.New("你已被移出群聊，无法发送消息")), conn)
			return
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

// singleSendAllowed 单聊发送鉴权：仅当确定"非好友"时拒绝；灰度关闭 / 缓存未知 / 回源失败一律放行（fail-open）。
func singleSendAllowed(srvCtx *svc.ServiceContext, senderUid, recvUid string) bool {
	if !srvCtx.Config.AuthzGate.Enabled || !srvCtx.Config.AuthzGate.SingleChat {
		return true // 灰度关闭 = 现状行为，不校验
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if srvCtx.RelationCache != nil {
		switch srvCtx.RelationCache.IsFriend(ctx, senderUid, recvUid) {
		case relationcache.VerdictAllowed:
			return true
		case relationcache.VerdictDenied:
			return false
		}
	}

	// Unknown：回源 FriendList 确认（顺带暖缓存）；回源失败 fail-open
	resp, err := srvCtx.Social.FriendList(ctx, &socialclient.FriendListReq{UserId: senderUid})
	if err != nil || resp == nil {
		return true
	}
	friendUids := make([]string, 0, len(resp.List))
	isFriend := false
	for _, f := range resp.List {
		friendUids = append(friendUids, f.FriendUid)
		if f.FriendUid == recvUid {
			isFriend = true
		}
	}
	if srvCtx.RelationCache != nil {
		// 回源版本用 0：好友删除走无条件 SREM-if-loaded，不依赖版本门
		_ = srvCtx.RelationCache.LoadFriends(ctx, senderUid, friendUids, 0)
	}
	return isFriend
}

// groupSendAllowed 群发送鉴权：仅当确定"非群成员"时拒绝；灰度关闭 / 缓存未知 / 回源失败一律放行（fail-open）。
func groupSendAllowed(srvCtx *svc.ServiceContext, senderUid, groupId string) bool {
	if !srvCtx.Config.AuthzGate.Enabled || !srvCtx.Config.AuthzGate.GroupChat {
		return true // 灰度关闭 = 现状行为，不校验
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if srvCtx.RelationCache != nil {
		switch srvCtx.RelationCache.IsGroupMember(ctx, groupId, senderUid) {
		case relationcache.VerdictAllowed:
			return true
		case relationcache.VerdictDenied:
			return false
		}
	}

	// Unknown：回源 GroupUsers 确认（顺带暖缓存）；回源失败 fail-open
	resp, err := srvCtx.Social.GroupUsers(ctx, &socialclient.GroupUsersReq{GroupId: groupId})
	if err != nil || resp == nil {
		return true
	}
	memberUids := make([]string, 0, len(resp.List))
	isMember := false
	for _, m := range resp.List {
		memberUids = append(memberUids, m.UserId)
		if m.UserId == senderUid {
			isMember = true
		}
	}
	if srvCtx.RelationCache != nil {
		// 回源版本用 0：与单聊一致，依赖 Kafka 按 gid 分区有序兜底
		_ = srvCtx.RelationCache.LoadGroupMembers(ctx, groupId, memberUids, 0)
	}
	return isMember
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
