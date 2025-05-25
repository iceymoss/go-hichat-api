package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Conversations 用户会话绑定，绑定用户所有会话记录
type Conversations struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	// 用户id
	UserId string `bson:"userId"`

	// 会话列表
	ConversationList map[string]*Conversation `bson:"conversationList"`

	UpdateAt time.Time `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
