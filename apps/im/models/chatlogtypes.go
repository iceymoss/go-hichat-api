package model

import (
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ChatLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`

	ConversationId string             `bson:"conversationId"`
	SendId         string             `bson:"sendId"`
	RecvId         string             `bson:"recvId"`
	MsgFrom        int                `bson:"msgFrom"`
	MsgType        constants.MType    `bson:"msgType"`
	MsgContent     string             `bson:"msgContent"`
	SendTime       int64              `bson:"sendTime"`
	Status         int                `bson:"status"`
	ChatType       constants.ChatType `bson:"chatType"`
	UpdateAt       time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt       time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
