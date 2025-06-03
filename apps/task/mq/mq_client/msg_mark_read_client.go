package mq_client

import (
	"context"
	"encoding/json"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/zeromicro/go-queue/kq"
)

type MsgReadTransferClient interface {
	Push(msg *mq.MsgMarkRead) error
}

type msgReadTransferClient struct {
	pusher *kq.Pusher
}

// NewMsgReadTransferClient 通过配置实例化客户端
func NewMsgReadTransferClient(addr []string, topic string, opts ...kq.PushOption) MsgReadTransferClient {
	return &msgReadTransferClient{
		pusher: kq.NewPusher(addr, topic),
	}
}

// Push 已读标记mq生产者
func (c *msgReadTransferClient) Push(msg *mq.MsgMarkRead) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	return c.pusher.Push(ctx, string(body))
}
