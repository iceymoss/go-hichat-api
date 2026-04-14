package ws

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type Msg struct {
	// 消息类型
	constants.MType `json:"mType" mapstructure:"mType"`

	// 消息内容
	Content string `json:"content" mapstructure:"content"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `json:"readRecords" mapstructure:"readRecords"`
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
