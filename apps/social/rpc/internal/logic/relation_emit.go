package logic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/relationcache"

	"gorm.io/gorm"
)

// emitRelationChange 在同一个 GORM 事务里执行关系变更(mutate)和 outbox 写入，原子提交，
// 保证"DB 变更 ⇔ 事件不丢"。提交后 best-effort 同步关系缓存（version = outbox.id），让闸门即时生效；
// 缓存失败被吞掉，由 relay 投递 + 消费端版本门 + TTL 最终自愈。
// payload 写入 outbox 时 Version 留空，relay 投递时用 outbox.id 回填。
func emitRelationChange(ctx context.Context, svcCtx *svc.ServiceContext, eventType, groupId string, ev *mq.RelationChangeTransfer, mutate func(tx *gorm.DB) error) error {
	ev.EventType = eventType
	if ev.Timestamp == 0 {
		ev.Timestamp = time.Now().UnixMilli()
	}

	tx := db.GetMysqlConn(db.MYSQL_DB_HICHAT2).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := mutate(tx); err != nil {
		tx.Rollback()
		return err
	}

	body, err := json.Marshal(ev)
	if err != nil {
		tx.Rollback()
		return err
	}
	row := &objects.RelationOutbox{EventType: eventType, GroupID: groupId, Payload: string(body), Status: 0}
	if err := svcCtx.RelationOutboxModel.InsertTx(tx, row); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	bestEffortApplyCache(ctx, svcCtx.RelationCache, ev, int64(row.ID))
	return nil
}

// emitRelationChangeInTx 在调用方已开启的事务 tx 内写入一条 outbox 事件（与关系变更原子提交）。
// 适用于成员新增等已自带 GORM 事务、不便交由 emitRelationChange 接管整段事务的入口。
// 不做提交后即时缓存同步：群扇出对亚秒级新鲜度不敏感，靠 relay 投递 + 消费端版本门 + TTL 最终一致即可。
func emitRelationChangeInTx(tx *gorm.DB, svcCtx *svc.ServiceContext, eventType, groupId string, ev *mq.RelationChangeTransfer) error {
	ev.EventType = eventType
	if ev.Timestamp == 0 {
		ev.Timestamp = time.Now().UnixMilli()
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	row := &objects.RelationOutbox{EventType: eventType, GroupID: groupId, Payload: string(body), Status: 0}
	return svcCtx.RelationOutboxModel.InsertTx(tx, row)
}

// bestEffortApplyCache 提交后把变更同步进关系缓存。与消费端 applyCache 同构，失败忽略。
func bestEffortApplyCache(ctx context.Context, rc *relationcache.Cache, ev *mq.RelationChangeTransfer, version int64) {
	if rc == nil {
		return
	}
	switch ev.EventType {
	case constants.RelationEventGroupMemberRemoved:
		_, _ = rc.RemoveGroupMember(ctx, ev.GroupId, ev.UserId, version)
	case constants.RelationEventGroupDisbanded:
		_ = rc.InvalidateGroup(ctx, ev.GroupId)
	case constants.RelationEventFriendDeleted:
		_, _ = rc.RemoveFriendIfLoaded(ctx, ev.FriendA, ev.FriendB)
		_, _ = rc.RemoveFriendIfLoaded(ctx, ev.FriendB, ev.FriendA)
	case constants.RelationEventFriendAdded:
		_, _ = rc.AddFriendIfLoaded(ctx, ev.FriendA, ev.FriendB)
		_, _ = rc.AddFriendIfLoaded(ctx, ev.FriendB, ev.FriendA)
	case constants.RelationEventGroupMemberAdded:
		_, _ = rc.AddGroupMember(ctx, ev.GroupId, ev.UserId, version)
	}
}
