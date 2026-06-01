package model

import (
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/constants"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	// TODO: Fill your own fields

	// 会话id
	ConversationId string `bson:"conversationId,omitempty"`

	// 聊天类型
	ChatType constants.ChatType `bson:"chatType,omitempty"`
	//TargetId       string             `bson:"targetId,omitempty"`

	// 是否展示（不能用 omitempty，否则 false 不会存储）
	IsShow bool `bson:"isShow"`

	// 会话置顶（不能用 omitempty，否则 false 不会存储）
	IsTop bool `bson:"isTop"`

	// 消息免打扰（不能用 omitempty，否则 false 不会存储）
	IsMute bool `bson:"isMute"`

	// 会话下消息总数（不能用 omitempty，否则 0 不会存储）
	Total int `bson:"total"`

	// 会话序号
	Seq int64 `bson:"seq"`

	// 当前会话的最晚的一条聊天记录，用于在会用会话聊天，展示给用户看最新的未读消息内容
	Msg *ChatLog `bson:"msg,omitempty"`

	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
