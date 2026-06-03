package msg_transfer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	userModels "github.com/iceymoss/go-hichat-api/apps/user/models"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/bitmap"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
)

// 用户设置里读已读回执开关的字段名，必须与前端 settings-store 保持一致
const readReceiptUserField = "readReceiptEnabled"

// MsgReadTransfer 实现ConsumeHandler接口
type MsgReadTransfer struct {
	*BaseChatTransfer
}

// NewMsgReadTransfer 实例化一个MsgReadTransfer
func NewMsgReadTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &MsgReadTransfer{
		BaseChatTransfer: NewBaseMsgChatTransfer(svc),
	}
}

func (m *MsgReadTransfer) Consume(ctx context.Context, key, value string) error {
	var (
		data mq.MsgMarkRead
		err  error
	)

	//1. 解析消息内容（如 JSON 反序列化）
	if err = json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	fmt.Printf("已经收到消息了用户已读标识: %+v\n", data)

	// 进会话/已读即清除该会话的 @我 角标（与回执开关无关，必须在 shouldDeliverReceipt 门禁之前执行）
	m.clearAtMe(ctx, data.ConversationId, data.SendId)

	// 三层开关判定（任一关闭 → 既不写 bitmap，也不推回执）
	// 关键：要在写 bitmap 之前拦截，否则"关闭方"的 bit 会永久留在 bitmap 里，
	// 后续其他成员读取触发新一轮 receipt 时会泄漏"关闭方已读"这个事实。
	if !m.shouldDeliverReceipt(ctx, data.SendId, data.RecvId, data.ChatType) {
		m.Infof("read receipt suppressed by switch: reader=%s, sender=%s, chatType=%v",
			data.SendId, data.RecvId, data.ChatType)
		return nil
	}

	// 写库更新 ReadRecords bitmap
	readRecords, err := m.UpdateChatLogRead(ctx, &data)
	if err != nil {
		return err
	}

	return m.MsgChatTransfer(ctx, &mq.MsgChatTransfer{
		ChatType:       data.ChatType,
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		// 关键修复：通过 MsgType 把 ContentMakeRead 暴露给 pusher → ws.Chat.Msg.MType，
		// 前端才能识别 mType===6 为已读回执。原先只设 ContentType（无 json tag）被吞掉。
		MsgType:     constants.ContentMakeRead,
		ContentType: constants.ContentMakeRead,
		ReadRecords: readRecords,
	})
}

// clearAtMe 清除某用户某会话的 @我 角标（reader 进会话/已读时调用）。无该会话条目或已是 false 时不写库。
func (m *MsgReadTransfer) clearAtMe(ctx context.Context, conversationId, readerId string) {
	if conversationId == "" || readerId == "" {
		return
	}
	conversations, err := m.svcCtx.ConversationsModel.FindByUserId(ctx, readerId)
	if err != nil || conversations == nil || conversations.ConversationList == nil {
		return
	}
	conv, ok := conversations.ConversationList[conversationId]
	if !ok || conv == nil || !conv.HasAtMe {
		return
	}
	conv.HasAtMe = false
	if _, err := m.svcCtx.ConversationsModel.Update(ctx, conversations); err != nil {
		m.Errorf("clearAtMe update failed: %v, reader=%s, conv=%s", err, readerId, conversationId)
	}
}

// shouldDeliverReceipt 三层开关：系统 ∧ 阅读者愿意回执 ∧ 原发送者愿意看到回执。
// 注意：单聊直接用两端用户；群聊仅判定系统 + 阅读者（发送者对群里所有人的意愿用自己的开关统一表达）。
func (m *MsgReadTransfer) shouldDeliverReceipt(ctx context.Context, readerId, origSenderId string, chatType constants.ChatType) bool {
	// 1) 系统开关
	if !m.svcCtx.SystemConfigModel.GetBool(ctx, userModels.SysCfgReadReceiptEnabled, true) {
		return false
	}

	// 2) 阅读者（reader，即本次 markRead 的发起者）
	if !readUserReadReceipt(ctx, m.svcCtx.UserSettingsModel, readerId) {
		return false
	}

	// 3) 原消息发送者：群聊里 origSenderId 是群ID，没有意义 → 跳过；私聊才判定
	if chatType == constants.SingleChatType {
		if !readUserReadReceipt(ctx, m.svcCtx.UserSettingsModel, origSenderId) {
			return false
		}
	}
	return true
}

// readUserReadReceipt 读取某个用户的"已读回执"开关，解析失败或无记录时视为开启（默认友好）
func readUserReadReceipt(ctx context.Context, model *userModels.UserSettingsModel, uid string) bool {
	if uid == "" {
		return true
	}
	n, err := strconv.ParseUint(uid, 10, 64)
	if err != nil || n == 0 {
		return true
	}
	return model.GetBoolField(ctx, n, readReceiptUserField, true)
}

func (m *MsgReadTransfer) UpdateChatLogRead(ctx context.Context, data *mq.MsgMarkRead) (map[string]string, error) {
	res := make(map[string]string)
	chatLogs, err := m.svcCtx.ChatLogModel.ListByMsgId(ctx, data.MsgIds)
	if err != nil {
		return res, err
	}

	m.Infof("chatLogs %v", chatLogs)

	readAt := time.Now().UnixNano()
	for _, chatlog := range chatLogs {
		switch data.ChatType {
		case constants.GroupChatType:
			readRecords := bitmap.Load(chatlog.ReadRecords)
			readRecords.Set(data.SendId)
			chatlog.ReadRecords = readRecords.Export()
		case constants.SingleChatType:
			chatlog.ReadRecords = []byte{1}
		}

		// 记录本次读者的已读时间（sender 自己的时间在 addChatLog 时已写入）
		if chatlog.ReadTimes == nil {
			chatlog.ReadTimes = map[string]int64{}
		}
		if _, exists := chatlog.ReadTimes[data.SendId]; !exists {
			chatlog.ReadTimes[data.SendId] = readAt
		}

		res[chatlog.ID.Hex()] = base64.StdEncoding.EncodeToString(chatlog.ReadRecords)

		err = m.svcCtx.ChatLogModel.UpdateMakeRead(ctx, chatlog.ID, chatlog.ReadRecords, chatlog.ReadTimes)
		if err != nil {
			m.Errorf("update make read err %v", err)
		}
	}

	return res, nil
}
