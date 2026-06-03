package model

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ ChatLogModel = (*customChatLogModel)(nil)

type (
	// ChatLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChatLogModel.
	ChatLogModel interface {
		chatLogModel
		// ListAfter 前向分页：返回 sendTime 严格大于 afterSendTime 的 count 条，按 sendTime 升序（用于"加载更新的消息"）
		ListAfter(ctx context.Context, conversationId string, afterSendTime, count int64) ([]*ChatLog, error)
	}

	customChatLogModel struct {
		*defaultChatLogModel
	}
)

// NewChatLogModel returns a model for the mongo.
func NewChatLogModel() ChatLogModel {
	mongoConn := db.GetMongoConn()
	return &customChatLogModel{
		defaultChatLogModel: newDefaultChatLogModel(mongoConn),
	}
}

// ListAfter 返回某会话中 sendTime 严格晚于 afterSendTime 的若干条消息，按时间升序排列。
// 与 ListBySendTime（降序、查更早）对称，供"向下滚动加载更新消息"使用。
func (m *customChatLogModel) ListAfter(ctx context.Context, conversationId string, afterSendTime, count int64) ([]*ChatLog, error) {
	if count <= 0 {
		count = DefaultChatLogCount
	}
	filter := bson.M{
		"conversationId": conversationId,
		"sendTime":       bson.M{"$gt": afterSendTime},
	}
	opt := options.Find().
		SetSort(bson.D{{Key: "sendTime", Value: 1}}).
		SetLimit(count)

	var data []*ChatLog
	cursor, err := m.Conn.Database(HiChat2).Collection(ChatLogs).Find(ctx, filter, opt)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &data); err != nil {
		return nil, fmt.Errorf("数据解码失败: %w", err)
	}
	return data, nil
}
