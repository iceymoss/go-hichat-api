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

var _ ConversationsModel = (*customConversationsModel)(nil)

type (
	// ConversationsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customConversationModel.
	ConversationsModel interface {
		conversationsModel
		EnsureGroup(context.Context, string, string, int64) (bool, error)
		SetGroupRemoved(context.Context, string, string, int64, int64) (bool, error)
		EnsureUniqueIndex(context.Context) error
		Update(context.Context, *Conversations) (*mongo.UpdateResult, error)
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

func (m *customConversationsModel) Update(ctx context.Context, data *Conversations) (*mongo.UpdateResult, error) {
	set := bson.M{"updateAt": time.Now()}
	for groupID, conversation := range data.ConversationList {
		if conversation == nil {
			continue
		}
		prefix := "conversationList." + groupID
		set[prefix+".conversationId"] = conversation.ConversationId
		set[prefix+".chatType"] = conversation.ChatType
		set[prefix+".isTop"] = conversation.IsTop
		set[prefix+".isMute"] = conversation.IsMute
		set[prefix+".hasAtMe"] = conversation.HasAtMe
		set[prefix+".total"] = conversation.Total
		set[prefix+".seq"] = conversation.Seq
		set[prefix+".msg"] = conversation.Msg
		set[prefix+".isShow"] = bson.M{"$cond": bson.A{bson.M{"$gt": bson.A{bson.M{"$ifNull": bson.A{"$" + prefix + ".removedAt", int64(0)}}, int64(0)}}, false, conversation.IsShow}}
	}
	return m.getColl().UpdateOne(ctx, bson.M{"userId": data.UserId}, mongo.Pipeline{{{Key: "$set", Value: set}}})
}

func (m *customConversationsModel) EnsureGroup(ctx context.Context, userID, groupID string, version int64) (bool, error) {
	now := time.Now()
	prefix := "conversationList." + groupID
	if version <= 0 {
		version = 1
	}
	existing := "$" + prefix
	condition := bson.M{"$lt": bson.A{bson.M{"$ifNull": bson.A{existing + ".relationVersion", int64(0)}}, version}}
	entry := bson.M{"$mergeObjects": bson.A{bson.M{"$ifNull": bson.A{existing, bson.M{}}}, bson.M{"conversationId": groupID, "chatType": constants.GroupChatType, "isShow": true, "removedAt": int64(0), "relationVersion": version}}}
	update := mongo.Pipeline{{{Key: "$set", Value: bson.M{"userId": bson.M{"$ifNull": bson.A{"$userId", userID}}, "createAt": bson.M{"$ifNull": bson.A{"$createAt", now}}, prefix: bson.M{"$cond": bson.A{condition, entry, existing}}}}}}
	result, err := m.getColl().UpdateOne(ctx, bson.M{"userId": userID}, update, options.Update().SetUpsert(true))
	return result != nil && (result.UpsertedCount == 1 || result.ModifiedCount == 1), err
}

func (m *customConversationsModel) SetGroupRemoved(ctx context.Context, userID, groupID string, removedAt, version int64) (bool, error) {
	prefix := "conversationList." + groupID
	now := time.Now()
	existing := "$" + prefix
	condition := bson.M{"$lt": bson.A{bson.M{"$ifNull": bson.A{existing + ".relationVersion", int64(0)}}, version}}
	entry := bson.M{"$mergeObjects": bson.A{bson.M{"$ifNull": bson.A{existing, bson.M{}}}, bson.M{"conversationId": groupID, "chatType": constants.GroupChatType, "isShow": false, "removedAt": removedAt, "relationVersion": version}}}
	update := mongo.Pipeline{{{Key: "$set", Value: bson.M{"userId": bson.M{"$ifNull": bson.A{"$userId", userID}}, "createAt": bson.M{"$ifNull": bson.A{"$createAt", now}}, prefix: bson.M{"$cond": bson.A{condition, entry, existing}}}}}}
	result, err := m.getColl().UpdateOne(ctx, bson.M{"userId": userID}, update, options.Update().SetUpsert(true))
	return result != nil && (result.UpsertedCount == 1 || result.ModifiedCount == 1), err
}

func (m *customConversationsModel) EnsureUniqueIndex(ctx context.Context) error {
	cursor, err := m.getColl().Aggregate(ctx, mongo.Pipeline{{{Key: "$group", Value: bson.M{"_id": "$userId", "count": bson.M{"$sum": 1}}}}, {{Key: "$match", Value: bson.M{"_id": bson.M{"$ne": ""}, "count": bson.M{"$gt": 1}}}}, {{Key: "$limit", Value: 1}}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		var duplicate bson.M
		if err := cursor.Decode(&duplicate); err != nil {
			return err
		}
		return fmt.Errorf("duplicate userId %v; run Mongo conversation cleanup before startup", duplicate["_id"])
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	_, err = m.getColl().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uk_user_id")})
	return err
}
