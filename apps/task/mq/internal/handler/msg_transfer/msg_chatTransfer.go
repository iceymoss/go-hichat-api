package msg_transfer

import (
	"context"
	"encoding/json"
	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
)

// MsgChatTransfer 实现ConsumeHandler接口
type MsgChatTransfer struct {
	logx.Logger
	svcCtx *svc.ServiceContext
}

// NewMsgChatTransfer 实例化一个MsgChatTransfer
func NewMsgChatTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	return &MsgChatTransfer{
		Logger: logx.WithContext(context.Background()),
		svcCtx: svc,
	}
}

func (m *MsgChatTransfer) Consume(ctx context.Context, key, value string) error {
	var (
		data mq.MsgChatTransfer
	)

	//1. 解析消息内容（如 JSON 反序列化）
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// 写入数据库（如 MongoDB 聊天记录）
	if err := m.addChatLog(ctx, data); err != nil {
		return err
	}

	//推送至 WebSocket 客户端
	err := m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data:      data,
	})

	//todo:
	// 4.错误处理（重试队列/DLQ）

	return err
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
	_, err := m.svcCtx.ChatLogModel.Insert(ctx, &chatLog)
	return err
}
