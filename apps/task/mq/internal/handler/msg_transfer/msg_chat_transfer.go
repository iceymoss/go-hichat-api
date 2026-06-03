package msg_transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/bitmap"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
)

// MsgChatTransfer 实现ConsumeHandler接口
type MsgChatTransfer struct {
	*BaseChatTransfer
}

// NewMsgChatTransfer 实例化一个MsgChatTransfer
func NewMsgChatTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &MsgChatTransfer{
		BaseChatTransfer: NewBaseMsgChatTransfer(svc),
	}
}

func (m *MsgChatTransfer) Consume(ctx context.Context, key, value string) error {
	var (
		data mq.MsgChatTransfer
		err  error
	)

	//1. 解析消息内容（如 JSON 反序列化）
	if err = json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	fmt.Printf("已经收到消息了: %+v\n", data)

	// 撤回控制帧：DB 已由 im-rpc 更新，这里不落库，仅复用通用 fan-out 推送给会话各端。
	if data.MsgType == constants.ContentRecall {
		return m.transferRecall(ctx, &data)
	}

	// 写入数据库（如 MongoDB 聊天记录）；同时把真实 MsgId 回填到 data，供下游推送
	if err = m.addChatLog(ctx, &data); err != nil {
		return err
	}

	// 1) 推送给接收方（原有行为）
	if err = m.MsgChatTransfer(ctx, &data); err != nil {
		return err
	}
	// 2) 额外回响给发送方：携带 MongoDB MsgId，前端把 local_ 占位 ID 升级为真实 ID
	return m.echoToSender(&data)
}

// transferRecall 推送撤回事件给会话各端：
// 先按会话成员扇出（group 跳过原发送者 / single 推给对端），再单独补推给原发送者本人（含其它在线设备）。
// 全程不改写 ContentType，保持前端能识别 ContentRecall。撤回幂等，重复推送无副作用。
func (m *MsgChatTransfer) transferRecall(ctx context.Context, data *mq.MsgChatTransfer) error {
	if err := m.MsgChatTransfer(ctx, data); err != nil {
		return err
	}
	if data.SendId == "" {
		return nil
	}
	// 补推给原发送者本人：走 single 路径（pusher 会 push 到 RecvId），不改 ContentType。
	echo := *data
	echo.RecvIdList = nil
	echo.RecvId = data.SendId
	echo.ChatType = constants.SingleChatType
	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      &echo,
	})
}

// addChatLog 将聊天记录消息持久化到数据库中
// 注意：入参改为指针，以便把 Insert 后生成的 ObjectID.Hex 回写到 data.MsgId
func (m *MsgChatTransfer) addChatLog(ctx context.Context, data *mq.MsgChatTransfer) error {
	// TODO(SendTime 单位不一致): 这里未透传 data.SendTime，落库时由 Insert 兜底成 UnixMilli（毫秒），
	// 而生产端 ws 把 data.SendTime 写成 UnixNano（纳秒），导致库里与推送两套单位。
	// 目前撤回逻辑已在 recallmsglogic.normalizeUnixMillis 做归一兜底，不影响功能；
	// 若要从源头统一，应在此显式赋 SendTime 并与 ws/前端的 sendTime 口径一起对齐（统一为毫秒）。
	chatLog := model.ChatLog{
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		MsgType:        data.MsgType,
		MsgContent:     data.MsgContent,
		Quote:          data.Quote,
		ChatType:       data.ChatType,
		AtUsers:        data.AtUsers,
		AtAll:          data.AtAll,
	}

	// 消息发起人标记为已读
	readRecords := bitmap.NewBitmap(0)
	readRecords.Set(chatLog.SendId)
	chatLog.ReadRecords = readRecords.Export()
	// 发送者自己的已读时间 = 发送时间
	chatLog.ReadTimes = map[string]int64{chatLog.SendId: time.Now().UnixNano()}

	// 记录聊天记录
	insertedID, err := m.svcCtx.ChatLogModel.Insert(ctx, &chatLog)
	if err != nil {
		zLog.Error("添加聊天记录失败", zap.Any("chatLog", chatLog), zap.Error(err))
		return err
	}

	// Insert 只在返回值里给出新生成的 ObjectID，不会回写 chatLog.ID。
	// 把这个 ID 同步到 chatLog 和 data，下游 pusher 会把它写进 WS Message.Id，
	// 前端 sender 侧据此把 local_ 占位 ID 升级为真实 mongoID。
	if insertedID != nil {
		chatLog.ID = *insertedID
		data.MsgId = insertedID.Hex()
	}

	// 更新会话的最后一次聊天记录
	err = m.svcCtx.ConversationModel.UpdateMsg(ctx, &chatLog)
	if err != nil {
		zLog.Error("更新会话失败", zap.Any("chatLog", chatLog), zap.Error(err))
		return err
	}

	// 发送者的消息自动标记为已读：user.Total += 1，防止发送者自己看到未读气泡
	conversations, err := m.svcCtx.ConversationsModel.FindByUserId(ctx, data.SendId)
	if err == nil && conversations.ConversationList != nil {
		if conv, ok := conversations.ConversationList[data.ConversationId]; ok {
			conv.Total = conv.Total + 1
			m.svcCtx.ConversationsModel.Update(ctx, conversations)
		}
	}

	// 群聊 @：给被 @的成员置会话级 HasAtMe 角标
	m.markAtMe(ctx, data)

	return nil
}

// markAtMe 为被 @的群成员置会话级 HasAtMe（进会话 / markRead 时清除）。
// @所有人时拉群成员（排除发送者）；@具体成员时用 atUsers。非群成员没有该会话条目，天然被跳过。
func (m *MsgChatTransfer) markAtMe(ctx context.Context, data *mq.MsgChatTransfer) {
	if data.ChatType != constants.GroupChatType || (!data.AtAll && len(data.AtUsers) == 0) {
		return
	}

	targets := make(map[string]struct{})
	if data.AtAll {
		res, err := m.svcCtx.Social.GroupUsers(ctx, &socialclient.GroupUsersReq{GroupId: data.RecvId})
		if err != nil {
			zLog.Error("markAtMe: group users failed", zap.Error(err), zap.String("groupId", data.RecvId))
			return
		}
		for _, mem := range res.List {
			if mem.UserId != data.SendId {
				targets[mem.UserId] = struct{}{}
			}
		}
	} else {
		for _, uid := range data.AtUsers {
			if uid != "" && uid != data.SendId {
				targets[uid] = struct{}{}
			}
		}
	}

	for uid := range targets {
		conversations, err := m.svcCtx.ConversationsModel.FindByUserId(ctx, uid)
		if err != nil || conversations == nil || conversations.ConversationList == nil {
			continue
		}
		conv, ok := conversations.ConversationList[data.ConversationId]
		if !ok || conv == nil {
			continue
		}
		if conv.HasAtMe {
			continue
		}
		conv.HasAtMe = true
		if _, err := m.svcCtx.ConversationsModel.Update(ctx, conversations); err != nil {
			zLog.Error("markAtMe: update conversations failed", zap.Error(err), zap.String("uid", uid))
		}
	}
}

func (m *MsgChatTransfer) single(ctx context.Context, data *mq.MsgChatTransfer) error {
	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})
}

func (m *MsgChatTransfer) group(ctx context.Context, data *mq.MsgChatTransfer) error {
	res, err := m.svcCtx.Social.GroupUsers(ctx, &socialclient.GroupUsersReq{
		GroupId: data.RecvId,
	})
	if err != nil {
		return err
	}

	data.RecvIdList = make([]string, 0, len(res.List))
	for _, member := range res.List {
		// 跳过发送人
		if member.UserId == data.SendId {
			continue
		}
		data.RecvIdList = append(data.RecvIdList, member.UserId)
	}

	return m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})
}
