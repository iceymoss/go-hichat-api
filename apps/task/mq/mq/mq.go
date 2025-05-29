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
	ReadRecords map[string]string `mapstructure:"readRecords"`

	// 消息内容
	MsgContent string `json:"content,omitempty"`
	// 发送时间
	SendTime int64 `json:"sendTime"`

	ContentType constants.MType
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
