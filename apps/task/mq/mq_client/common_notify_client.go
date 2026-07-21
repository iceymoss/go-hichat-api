package mq_client

import (
	"context"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"

	"github.com/zeromicro/go-queue/kq"
)

// CommonNotifyClient 公共通知通道生产者。业务方处理完动作后直接 Push（参考撤回范式）。
type CommonNotifyClient interface {
	Push(msg *mq.CommonNotify) error
}

type commonNotifyClient struct {
	pusher *kq.Pusher
}

// NewCommonNotifyClient 通过配置实例化公共通知事件生产者
func NewCommonNotifyClient(addr []string, topic string, opts ...kq.PushOption) CommonNotifyClient {
	return &commonNotifyClient{
		pusher: kq.NewPusher(addr, topic, opts...),
	}
}

// Push 将公共通知事件 push 到 kafka commonNotifyTransfer topic
func (c *commonNotifyClient) Push(msg *mq.CommonNotify) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.pusher.Push(context.Background(), string(body))
}
