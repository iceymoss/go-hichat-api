package model

import (
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"time"

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

	// 是否展示
	IsShow bool `bson:"isShow,omitempty"`

	// 会话下消息总数
	Total int `bson:"total,omitempty"`

	Seq int64 `bson:"seq"`

	// 会话消息
	Msg *ChatLog `bson:"msg,omitempty"`

	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
