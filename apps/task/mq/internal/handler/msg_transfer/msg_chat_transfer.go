package msg_transfer

import (
	"context"
	"encoding/json"
	"fmt"
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

	// 写入数据库（如 MongoDB 聊天记录）
	if err = m.addChatLog(ctx, data); err != nil {
		return err
	}

	return m.MsgChatTransfer(ctx, &data)
}

// addChatLog 将聊天记录消息持久化到数据库中
func (m *MsgChatTransfer) addChatLog(ctx context.Context, data mq.MsgChatTransfer) error {
	chatLog := model.ChatLog{
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		MsgType:        data.MsgType,
		MsgContent:     data.MsgContent,
		ChatType:       data.ChatType,
	}

	// 消息发起人标记为已读
	readRecords := bitmap.NewBitmap(0)
	readRecords.Set(chatLog.SendId)
	chatLog.ReadRecords = readRecords.Export()

	// 记录聊天记录
	_, err := m.svcCtx.ChatLogModel.Insert(ctx, &chatLog)
	if err != nil {
		zLog.Error("添加聊天记录失败", zap.Any("chatLog", chatLog), zap.Error(err))
		return err
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

	return nil
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
