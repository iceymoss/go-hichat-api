package model

import (
	"context"
	"fmt"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		// UpdateRecalled 条件撤回：仅当消息当前为正常态时置为已撤回，返回本次是否真正改动（用于幂等区分是否需要推送）
		UpdateRecalled(ctx context.Context, id primitive.ObjectID, recalledBy string, recalledAt int64) (bool, error)
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

// UpdateRecalled 条件撤回：filter 限定 status=正常，避免重复撤回 / 并发撤回时多次置态。
// 返回 recalled=true 表示本次调用真正把消息从正常翻为已撤回（调用方据此决定是否推送撤回事件，保证幂等）。
func (m *customChatLogModel) UpdateRecalled(ctx context.Context, id primitive.ObjectID, recalledBy string, recalledAt int64) (bool, error) {
	if id.IsZero() {
		return false, fmt.Errorf("无效的消息ID")
	}
	filter := bson.M{"_id": id, "status": constants.MsgStatusNormal}
	update := bson.M{"$set": bson.M{
		"status":     constants.MsgStatusRecalled,
		"recalledBy": recalledBy,
		"recalledAt": recalledAt,
		"updateAt":   time.Now(),
	}}
	res, err := m.Conn.Database(HiChat2).Collection(ChatLogs).UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("数据库更新失败: %w", err)
	}
	return res.ModifiedCount > 0, nil
}
