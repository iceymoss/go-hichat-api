package model

import (
	"context"
	"fmt"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var _ ConversationModel = (*customConversationModel)(nil)

type (
	// ConversationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConversationModel.
	ConversationModel interface {
		conversationModel
		EnsureGroup(context.Context, string) (bool, error)
		EnsureUniqueIndex(context.Context) error
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

func (m *customConversationModel) EnsureGroup(ctx context.Context, groupID string) (bool, error) {
	now := time.Now()
	result, err := m.getColl().UpdateOne(ctx, bson.M{"conversationId": groupID}, bson.M{"$setOnInsert": bson.M{"conversationId": groupID, "chatType": constants.GroupChatType, "isShow": true, "isTop": false, "isMute": false, "hasAtMe": false, "total": 0, "seq": 0, "createAt": now, "updateAt": now}}, options.Update().SetUpsert(true))
	return result != nil && result.UpsertedCount == 1, err
}

func (m *customConversationModel) EnsureUniqueIndex(ctx context.Context) error {
	cursor, err := m.getColl().Aggregate(ctx, mongo.Pipeline{{{Key: "$group", Value: bson.M{"_id": "$conversationId", "count": bson.M{"$sum": 1}}}}, {{Key: "$match", Value: bson.M{"_id": bson.M{"$ne": ""}, "count": bson.M{"$gt": 1}}}}, {{Key: "$limit", Value: 1}}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		var duplicate bson.M
		if err := cursor.Decode(&duplicate); err != nil {
			return err
		}
		return fmt.Errorf("duplicate conversationId %v; run Mongo conversation cleanup before startup", duplicate["_id"])
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	_, err = m.getColl().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "conversationId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uk_conversation_id")})
	return err
}
