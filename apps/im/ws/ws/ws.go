package ws

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type Msg struct {
	// 消息类型
	constants.MType `json:"mType" mapstructure:"mType"`

	// 消息内容
	Content string `json:"content" mapstructure:"content"`

	// 引用/回复消息（JSON 字符串：{"id","name","preview"}），空表示非引用消息
	Quote string `json:"quote,omitempty" mapstructure:"quote"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`

	// 被 @ 的成员 uid 列表（仅群聊）
	AtUsers []string `json:"atUsers,omitempty" mapstructure:"atUsers"`

	// 是否 @所有人（仅群聊）
	AtAll bool `json:"atAll,omitempty" mapstructure:"atAll"`
}

// Chat 聊天会话， message 结构中的data字段
type Chat struct {
	// 会话id
	ConversationId string `json:"conversationId" mapstructure:"conversationId"`

	// 聊天类型
	constants.ChatType `json:"chatType" mapstructure:"chatType"`

	// 发送者
	SendId string `json:"sendId" mapstructure:"sendId"`

	// 接收者
	RecvId string `json:"recvId" mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `json:"sendTime" mapstructure:"sendTime"`

	// 附加内容类型（如 ContentMakeRead=6 已读回执, ContentMsgAck=7 发送方回响, ContentRecall=9 撤回），0 表示普通消息
	ContentType constants.MType `json:"contentType,omitempty" mapstructure:"contentType"`

	// 撤回操作者 uid，仅 ContentType=ContentRecall 时有值，前端据此区分"你/对方/管理员撤回了一条消息"
	RecalledBy string `json:"recalledBy,omitempty" mapstructure:"recalledBy"`

	// 发送内容
	Msg `json:"msg" mapstructure:"msg"`
}

type Push struct {
	// 会话id
	ConversationId string `json:"conversationId" mapstructure:"conversationId"`

	// 聊天类型：1. 私聊、2. 群聊
	constants.ChatType `json:"chatType" mapstructure:"chatType"`

	// 发送者
	SendId string `json:"sendId" mapstructure:"sendId"`

	RecvIdList []string `json:"recvIdList" mapstructure:"recvIdList"`

	// 接收者
	RecvId string `json:"recvId" mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `json:"sendTime" mapstructure:"sendTime"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`

	constants.MType `json:"mType" mapstructure:"mType"`
	Content         string `json:"content" mapstructure:"content"`

	// 引用/回复消息（JSON 字符串），随推送转发给接收方
	Quote string `json:"quote,omitempty" mapstructure:"quote"`

	// MongoDB 聊天记录 ObjectID（hex），用于前端把 local_ 占位 ID 替换为真实 ID
	MsgId string `json:"msgId" mapstructure:"msgId"`
	// 附加内容类型（ContentType=7 表示给发送方的回响）
	ContentType constants.MType `json:"contentType" mapstructure:"contentType"`

	// 撤回操作者 uid，仅 ContentType=ContentRecall 时有值
	RecalledBy string `json:"recalledBy,omitempty" mapstructure:"recalledBy"`

	// 被 @ 的成员 uid 列表（仅群聊普通消息）
	AtUsers []string `json:"atUsers,omitempty" mapstructure:"atUsers"`

	// 是否 @所有人（仅群聊普通消息）
	AtAll bool `json:"atAll,omitempty" mapstructure:"atAll"`
}

// MarkRead 已读标记
type MarkRead struct {
	constants.ChatType `json:"chatType" mapstructure:"chatType"`
	RecvId             string `json:"recvId" mapstructure:"recvId"`
	ConversationId     string `json:"conversationId" mapstructure:"conversationId"`
	// 发送者
	SendId      string            `json:"sendId" mapstructure:"sendId"`
	MsgIds      []string          `json:"msgIds" mapstructure:"msgIds"`
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`
}
