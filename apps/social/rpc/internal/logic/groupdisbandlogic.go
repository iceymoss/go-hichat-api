package logic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupDisbandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupDisbandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupDisbandLogic {
	return &GroupDisbandLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupDisbandLogic) GroupDisband(in *social.GroupDisbandReq) (*social.GroupDisbandResp, error) {
	// 1) 校验群存在且操作者为群主
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found %v", in.GroupId)
	}
	if group.CreatorUid != in.UserId {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only owner can disband group")
	}

	// 已解散则幂等成功
	if group.Status == 1 {
		return &social.GroupDisbandResp{}, nil
	}

	now := time.Now()
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()

	// 2) 标记群状态=已解散
	group.Status = 1
	group.UpdatedAt = now
	if err := tx.Table("groups").Where("id = ?", group.Id).Updates(map[string]any{
		"status":     group.Status,
		"updated_at": group.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group status err")
	}

	// 3) 清理成员（使群不可再访问）
	if err := tx.Table("group_members").Where("group_id = ?", in.GroupId).Delete(nil).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "delete group members err")
	}

	// 4) 清理申请（可选，不影响幂等）
	_ = tx.Table("group_requests").Where("group_id = ?", in.GroupId).Delete(nil).Error

	// 5) 同事务写 outbox（群解散事件），与上述变更原子提交
	ev := &mq.RelationChangeTransfer{
		EventType:  constants.RelationEventGroupDisbanded,
		GroupId:    in.GroupId,
		OperatorId: in.UserId,
		Timestamp:  now.UnixMilli(),
	}
	body, mErr := json.Marshal(ev)
	if mErr != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "marshal disband outbox err")
	}
	outboxRow := &objects.RelationOutbox{EventType: ev.EventType, GroupID: in.GroupId, Payload: string(body), Status: 0}
	if err := l.svcCtx.RelationOutboxModel.InsertTx(tx, outboxRow); err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "insert disband outbox err")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "commit disband err")
	}

	// 提交后 best-effort 失效整群缓存（失败由 relay + TTL 自愈）
	bestEffortApplyCache(l.ctx, l.svcCtx.RelationCache, ev, int64(outboxRow.ID))

	return &social.GroupDisbandResp{}, nil
}
