package mq

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type MsgChatTransfer struct {
	// 消息类型：1. 私聊、2. 群聊
	ChatType constants.ChatType `json:"chatType"`
	// 会话id
	ConversationId string `json:"conversationId"`
	// 发送者
	SendId string `json:"sendId"`
	// 接收者
	RecvId string `json:"recvId"`

	RecvIdList []string `json:"recvIdList"`

	// 消息类型
	MsgType constants.MType `json:"mType,omitempty"`

	// 已读记录
	ReadRecords map[string]string `json:"readRecords"`

	// 消息内容
	MsgContent string `json:"content,omitempty"`

	// 引用/回复消息（JSON 字符串：{"id","name","preview"}）
	Quote string `json:"quote,omitempty"`

	// 发送时间
	SendTime int64 `json:"sendTime"`

	ContentType constants.MType

	// MongoDB 持久化后的聊天记录 ID（hex），用于前端把 local_ 占位 ID 替换为真实 ID
	MsgId string `json:"msgId,omitempty"`

	// RecalledBy 撤回操作者 uid，仅 MsgType=ContentRecall 的撤回控制帧填充
	RecalledBy string `json:"recalledBy,omitempty"`

	// AtUsers 被 @ 的成员 uid 列表（仅群聊普通消息）
	AtUsers []string `json:"atUsers,omitempty"`

	// AtAll 是否 @所有人（仅群聊普通消息，权限已在生产端把关）
	AtAll bool `json:"atAll,omitempty"`
}

// MsgMarkRead 标记已读消息结构体
type MsgMarkRead struct {
	// 消息类型：1. 私聊、2. 群聊
	ChatType constants.ChatType `json:"chatType,omitempty"`
	// 会话id
	ConversationId string `json:"conversationId,omitempty"`
	// 发送者
	SendId string `json:"sendId,omitempty"`
	// 接收者
	RecvId string `json:"recvId,omitempty"`

	// 已读消息集合
	MsgIds []string `json:"msgIds,omitempty"`

	ReadRecords map[string]string `mapstructure:"readRecords"`
}
