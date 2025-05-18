package model

import (
	"github.com/iceymoss/go-hichat-api/pkg/db"
)

var _ ConversationModel = (*customConversationModel)(nil)

type (
	// ConversationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConversationModel.
	ConversationModel interface {
		conversationModel
	}

	customConversationModel struct {
		*defaultConversationModel
	}
)

// NewConversationModel returns a model for the mongo.
func NewConversationModel() ConversationModel {
	conn := db.GetMongoConn()
	return &customConversationModel{
		defaultConversationModel: newDefaultConversationModel(conn),
	}
}
