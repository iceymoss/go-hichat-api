package msg_transfer

import (
	"context"
	"encoding/json"
	"fmt"
	model "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

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

	fmt.Printf("已经收到消息了: %+v\n", data)

	// 写入数据库（如 MongoDB 聊天记录）
	if err := m.addChatLog(ctx, data); err != nil {
		return err
	}

	//推送至 WebSocket 客户端
	err := m.svcCtx.WsClient.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push",
		FormId:    constants.REDIS_SYSTEM_ROOT_TOEKN,
		Data:      data,
	})
	if err != nil {
		zLog.Error("Consume.Send: push to websocket serve failed", zap.Any("msg", data), zap.Error(err))
	}

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

	// 记录聊天记录
	_, err := m.svcCtx.ChatLogModel.Insert(ctx, &chatLog)

	// 更新会话的最后一次聊天记录
	err = m.svcCtx.ConversationModel.UpdateMsg(ctx, &chatLog)

	return err
}
