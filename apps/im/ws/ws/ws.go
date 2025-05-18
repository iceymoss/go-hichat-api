package ws

import "github.com/iceymoss/go-hichat-api/pkg/constants"

type Msg struct {
	constants.MType `mapstructure:"mType"`
	Content         string `mapstructure:"content"`
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

	// 接收者
	RecvId string `mapstructure:"recvId"`

	// 发送时间
	SendTime int64 `mapstructure:"sendTime"`

	constants.MType `mapstructure:"mType"`
	Content         string `mapstructure:"content"`
}
