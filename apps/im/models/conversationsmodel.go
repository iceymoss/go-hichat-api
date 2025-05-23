package model

import (
	"github.com/iceymoss/go-hichat-api/pkg/db"
)

var _ ConversationsModel = (*customConversationsModel)(nil)

type (
	// ConversationsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConversationModel.
	ConversationsModel interface {
		conversationsModel
	}

	customConversationsModel struct {
		*defaultConversationsModel
	}
)

// NewConversationsModel returns a model for the mongo.
func NewConversationsModel() ConversationsModel {
	conn := db.GetMongoConn()
	return &customConversationsModel{
		defaultConversationsModel: newDefaultConversationsModel(conn),
	}
}
