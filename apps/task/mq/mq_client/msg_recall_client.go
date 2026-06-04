package mq_client

import (
	"context"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"

	"github.com/zeromicro/go-queue/kq"
)

type MsgRecallTransferClient interface {
	Push(msg *mq.MsgRecallTransfer) error
}

type msgRecallTransferClient struct {
	pusher *kq.Pusher
}

// NewMsgRecallTransferClient 通过配置实例化撤回事件生产者
func NewMsgRecallTransferClient(addr []string, topic string, opts ...kq.PushOption) MsgRecallTransferClient {
	return &msgRecallTransferClient{
		pusher: kq.NewPusher(addr, topic, opts...),
	}
}

// Push 将撤回事件 push 到 kafka 中
func (c *msgRecallTransferClient) Push(msg *mq.MsgRecallTransfer) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	return c.pusher.Push(ctx, string(body))
}
