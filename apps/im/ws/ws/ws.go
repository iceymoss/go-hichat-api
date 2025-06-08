package ws

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type Msg struct {
	// 消息类型
	constants.MType `mapstructure:"mType"`

	// 消息内容
	Content string `mapstructure:"content"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `mapstructure:"readRecords"`
}

// Chat 聊天会话， message 结构中的data字段
type Chat struct {
	// 会话id
	ConversationId string `mapstructure:"conversationId"`

	// 聊天类型
	constants.ChatType `mapstructure:"chatType"`

	// 发送者
	SendId string `mapstructure:"sendId"`

	// 接收者
	RecvId string `mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `mapstructure:"sendTime"`

	// 发送内容
	Msg `mapstructure:"msg"`
}

type Push struct {
	// 会话id
	ConversationId string `mapstructure:"conversationId"`

	// 聊天类型：1. 私聊、2. 群聊
	constants.ChatType `mapstructure:"chatType"`

	// 发送者
	SendId string `mapstructure:"sendId"`

	RecvIdList []string `mapstructure:"recvIdList"`

	// 接收者
	RecvId string `mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `mapstructure:"sendTime"`

	// 已读记录 key为消息id、value为消息值
	ReadRecords map[string]string `mapstructure:"readRecords"`

	constants.MType `mapstructure:"mType"`
	Content         string `mapstructure:"content"`
}

// MarkRead 已读标记
type MarkRead struct {
	constants.ChatType `mapstructure:"chatType"`
	RecvId             string `mapstructure:"recvId"`
	ConversationId     string `mapstructure:"conversationId"`
	// 发送者
	SendId      string            `mapstructure:"sendId"`
	MsgIds      []string          `mapstructure:"msgIds"`
	ReadRecords map[string]string `mapstructure:"readRecords"`
}
