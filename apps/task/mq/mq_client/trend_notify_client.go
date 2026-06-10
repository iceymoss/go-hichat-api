package mq_client

import (
	"context"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"

	"github.com/zeromicro/go-queue/kq"
)

type TrendNotifyTransferClient interface {
	Push(msg *mq.TrendNotifyTransfer) error
}

type trendNotifyTransferClient struct {
	pusher *kq.Pusher
}

// NewTrendNotifyTransferClient 通过配置实例化动态消息通知事件生产者
func NewTrendNotifyTransferClient(addr []string, topic string, opts ...kq.PushOption) TrendNotifyTransferClient {
	return &trendNotifyTransferClient{
		pusher: kq.NewPusher(addr, topic, opts...),
	}
}

// Push 将动态消息通知事件 push 到 kafka 中
func (c *trendNotifyTransferClient) Push(msg *mq.TrendNotifyTransfer) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.pusher.Push(context.Background(), string(body))
}
