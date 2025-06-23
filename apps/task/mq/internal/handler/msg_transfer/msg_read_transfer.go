package msg_transfer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/bitmap"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
)

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
	// 业务处理 -- 更新
	readRecords, err := m.UpdateChatLogRead(ctx, &data)
	if err != nil {
		return err
	}

	return m.MsgChatTransfer(ctx, &mq.MsgChatTransfer{
		ChatType:       data.ChatType,
		ConversationId: data.ConversationId,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ContentType:    constants.ContentMakeRead,
		ReadRecords:    readRecords,
	})
}

func (m *MsgReadTransfer) UpdateChatLogRead(ctx context.Context, data *mq.MsgMarkRead) (map[string]string, error) {
	res := make(map[string]string)
	chatLogs, err := m.svcCtx.ChatLogModel.ListByMsgId(ctx, data.MsgIds)
	if err != nil {
		return res, err
	}

	m.Infof("chatLogs %v", chatLogs)

	for _, chatlog := range chatLogs {
		switch data.ChatType {
		case constants.GroupChatType:
			readRecords := bitmap.Load(chatlog.ReadRecords)
			readRecords.Set(data.SendId)
			chatlog.ReadRecords = readRecords.Export()
		case constants.SingleChatType:
			chatlog.ReadRecords = []byte{1}
		}

		res[chatlog.ID.Hex()] = base64.StdEncoding.EncodeToString(chatlog.ReadRecords)

		err = m.svcCtx.ChatLogModel.UpdateMakeRead(ctx, chatlog.ID, chatlog.ReadRecords)
		if err != nil {
			m.Errorf("update make read err %v", err)
		}
	}

	return res, nil
}
