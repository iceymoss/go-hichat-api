package mq_client

import (
	"context"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"github.com/zeromicro/go-queue/kq"
)

type RelationChangeTransferClient interface {
	Push(msg *mq.RelationChangeTransfer) error
}

type relationChangeTransferClient struct {
	pusher *kq.Pusher
}

// NewRelationChangeTransferClient 实例化关系变更事件生产者（relay 用）。
func NewRelationChangeTransferClient(addr []string, topic string, opts ...kq.PushOption) RelationChangeTransferClient {
	return &relationChangeTransferClient{
		pusher: kq.NewPusher(addr, topic, opts...),
	}
}

// Push 将关系变更事件投递到 kafka。按实体分区 key 投递，保证同群/同好友对事件有序。
func (c *relationChangeTransferClient) Push(msg *mq.RelationChangeTransfer) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.pusher.PushWithKey(context.Background(), partitionKey(msg), string(body))
}

// partitionKey 同群事件用 groupId，好友事件用有序好友对，确保同实体落同分区、按序消费。
// 好友的新增与删除事件落同一好友对分区 -> 天然有序，避免"删除后又被旧新增复活"。
func partitionKey(msg *mq.RelationChangeTransfer) string {
	switch msg.EventType {
	case constants.RelationEventFriendDeleted, constants.RelationEventFriendAdded:
		a, b := msg.FriendA, msg.FriendB
		if a > b {
			a, b = b, a
		}
		return "frd:" + a + ":" + b
	}
	return "grp:" + msg.GroupId
}
