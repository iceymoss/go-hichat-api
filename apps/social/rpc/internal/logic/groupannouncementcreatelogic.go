package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupAnnouncementCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupAnnouncementCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAnnouncementCreateLogic {
	return &GroupAnnouncementCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 公告历史/置顶
func (l *GroupAnnouncementCreateLogic) GroupAnnouncementCreate(in *social.GroupAnnouncementCreateReq) (*social.GroupAnnouncementCreateResp, error) {
	if in.Content == "" {
		return nil, errors.Wrapf(xerr.NewMsg("content required"), "content required")
	}
	if len(in.Content) > 1024 {
		return nil, errors.Wrapf(xerr.NewMsg("content too long"), "content too long")
	}

	// 权限：管理员/群主
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}
	if member.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only admin/owner can publish")
	}

	groupIdInt, err := strconv.Atoi(in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}
	uidInt, err := strconv.ParseUint(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid userId"), "invalid userId")
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()
	now := time.Now()

	// 插入公告
	res := tx.Table("group_announcements").Create(map[string]any{
		"group_id":   groupIdInt,
		"content":    in.Content,
		"created_by": uidInt,
		"created_at": now,
		"pinned":     0,
		"pinned_at":  nil,
		"deleted":    0,
	})
	if res.Error != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "create announcement err")
	}
	// 获取 last insert id（gorm map create 不会回填）
	var insertedId int64
	if err := tx.Raw("SELECT LAST_INSERT_ID()").Scan(&insertedId).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "get insert id err")
	}

	// 同步 groups.notification（展示用：默认展示最新公告）
	if err := tx.Table("groups").Where("id = ?", groupIdInt).
		Updates(map[string]any{"notification": in.Content, "notification_uid": uidInt, "updated_at": now}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group notification err")
	}

	tx.Commit()
	return &social.GroupAnnouncementCreateResp{Id: insertedId}, nil
}
