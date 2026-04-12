package socialmodels

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var _ FriendRequestsModel = (*customFriendRequestsModel)(nil)

type (
	// FriendRequestsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFriendRequestsModel.
	FriendRequestsModel interface {
		friendRequestsModel
		CountMessageRequests(ctx context.Context, userId string) (int64, error)
		InvalidateCountCache(ctx context.Context, userIds ...string)
		SoftDelete(ctx context.Context, id uint64, userId string) error
		IgnoreOtherPending(ctx context.Context, excludeId uint64, userIdA, userIdB uint64) error
	}

	customFriendRequestsModel struct {
		*defaultFriendRequestsModel
	}
)

// NewFriendRequestsModel returns a model for the database table.
func NewFriendRequestsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) FriendRequestsModel {
	return &customFriendRequestsModel{
		defaultFriendRequestsModel: newFriendRequestsModel(conn, c, opts...),
	}
}

// SoftDelete 软删除好友申请记录（将 status 设为 0）
func (m *customFriendRequestsModel) SoftDelete(ctx context.Context, id uint64, userId string) error {
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	res := mysqlConn.Table(m.table).
		Where("id = ?", id).
		Where("user_id = ? OR req_uid = ?", userId, userId).
		Update("status", 0)
	return res.Error
}

// IgnoreOtherPending 将同一对用户之间除 excludeId 外的所有待处理申请标记为已忽略
// A->B 或 B->A 的待处理申请全部忽略
func (m *customFriendRequestsModel) IgnoreOtherPending(ctx context.Context, excludeId uint64, userIdA, userIdB uint64) error {
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	res := mysqlConn.Table(m.table).
		Where("id != ?", excludeId).
		Where("handle_result = ?", 0). // 只处理待处理的
		Where("status = ?", 1).        // 未删除的
		Where(
			"(user_id = ? AND req_uid = ?) OR (user_id = ? AND req_uid = ?)",
			userIdA, userIdB, userIdB, userIdA,
		).
		Updates(map[string]interface{}{
			"handle_result": 3, // 已忽略
		})
	return res.Error
}

const friendReqCountCachePrefix = "friend_req_count:"
const friendReqCountCacheTTL = 60 * time.Second // 1 分钟缓存

// countCacheKey 生成用户气泡数的缓存 key
func countCacheKey(userId string) string {
	return friendReqCountCachePrefix + userId
}

// InvalidateCountCache 使指定用户的气泡数缓存失效
func (m *customFriendRequestsModel) InvalidateCountCache(ctx context.Context, userIds ...string) {
	rdb := db.GetRedisConn()
	for _, uid := range userIds {
		rdb.Del(ctx, countCacheKey(uid))
	}
}

// CountMessageRequests 统计好友申请气泡数（带 Redis 缓存，1 分钟 TTL）
func (m *customFriendRequestsModel) CountMessageRequests(ctx context.Context, userId string) (int64, error) {
	rdb := db.GetRedisConn()
	cacheKey := countCacheKey(userId)

	// 尝试读缓存
	cached, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		if v, e := strconv.ParseInt(cached, 10, 64); e == nil {
			return v, nil
		}
	}

	// 缓存未命中，查库
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)

	// 1. 我收到的待处理，接收方未读
	var receivedCount int64
	err = mysqlConn.Table(m.table).
		Where("req_uid = ?", userId).
		Where("handle_result = ?", 0).
		Where("status = ?", 1).
		Where("receiver_read = ?", 0).
		Count(&receivedCount).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	// 2. 我发出的有结果(同意/拒绝)，发起方未读
	var resultCount int64
	err = mysqlConn.Table(m.table).
		Where("user_id = ?", userId).
		Where("handle_result IN ?", []int{1, 2}).
		Where("status = ?", 1).
		Where("sender_read = ?", 0).
		Count(&resultCount).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	total := receivedCount + resultCount

	// 写入缓存
	rdb.Set(ctx, cacheKey, fmt.Sprintf("%d", total), friendReqCountCacheTTL)

	return total, nil
}
