package mq_client

import (
	"context"
	"encoding/json"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/zeromicro/go-queue/kq"
)

type MsgChatTransferClient interface {
	Push(msg *mq.MsgChatTransfer) error
}

type MsgChatTransferClientImpl struct {
	pusher *kq.Pusher
}

func NewMsgChatTransferClient(addrs []string, topic string, opts ...kq.PushOption) MsgChatTransferClient {
	client := &MsgChatTransferClientImpl{
		pusher: kq.NewPusher(addrs, topic, opts...),
	}
	return client
}

// Push 将消息push到kafka中
func (c *MsgChatTransferClientImpl) Push(msg *mq.MsgChatTransfer) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	return c.pusher.Push(ctx, string(body))
}
