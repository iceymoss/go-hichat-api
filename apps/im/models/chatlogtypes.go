package model

import (
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatLog 聊天记录
type ChatLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// 会话id
	ConversationId string `bson:"conversationId"`

	// 发送者id
	SendId string `bson:"sendId"`

	// 接受者id
	RecvId string `bson:"recvId"`

	// 消息来源
	MsgFrom int `bson:"msgFrom"`

	// 消息类型
	MsgType constants.MType `bson:"msgType"`

	// 消息内容
	MsgContent string `bson:"msgContent"`

	// 发送时间
	SendTime int64 `bson:"sendTime"`

	// 消息状态
	Status int `bson:"status"`

	// 聊天类型
	ChatType constants.ChatType `bson:"chatType"`

	// 已读记录
	ReadRecords []byte `bson:"readRecords"`

	// 每个成员的已读时间戳（unix nano），key=userId。群聊里用来展示"某成员何时读了"
	// sender 自己由 addChatLog 记为发送即读；其他成员由 msg_read_transfer 填入。
	ReadTimes map[string]int64 `bson:"readTimes,omitempty" json:"readTimes,omitempty"`

	// 更新时间
	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`

	// 创建时间
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
