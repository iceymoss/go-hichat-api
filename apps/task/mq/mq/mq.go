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

// MsgRecallTransfer 撤回事件消息结构体。
// 撤回与普通聊天消息是两类语义不同的领域事件，走独立 topic（msgRecallTransfer），
// 消费端据此扇出撤回控制帧，避免与聊天消息共享通道造成的领域耦合与吞吐相互影响。
// 注意：仅承载撤回所需字段；推送给前端的帧仍由消费端翻译成 ContentRecall 控制帧，对外协议不变。
type MsgRecallTransfer struct {
	// 消息类型：1. 私聊、2. 群聊
	ChatType constants.ChatType `json:"chatType"`
	// 会话id
	ConversationId string `json:"conversationId"`
	// 原消息发送者
	SendId string `json:"sendId"`
	// 原消息接收者（私聊为对端 uid，群聊为群 id）
	RecvId string `json:"recvId"`
	// 被撤回的消息 ID
	MsgId string `json:"msgId"`
	// RecalledBy 撤回操作者 uid
	RecalledBy string `json:"recalledBy"`
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
