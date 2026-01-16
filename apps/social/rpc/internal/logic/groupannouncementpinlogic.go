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

type GroupAnnouncementPinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupAnnouncementPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAnnouncementPinLogic {
	return &GroupAnnouncementPinLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupAnnouncementPinLogic) GroupAnnouncementPin(in *social.GroupAnnouncementPinReq) (*social.GroupAnnouncementPinResp, error) {
	// 权限：管理员/群主
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}
	if member.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only admin/owner can pin")
	}

	groupIdInt, err := strconv.Atoi(in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()
	now := time.Now()

	// 置顶：同一群只允许一个 pinned
	if in.Pinned {
		if err := tx.Table("group_announcements").
			Where("group_id = ? AND pinned = 1", groupIdInt).
			Updates(map[string]any{"pinned": 0, "pinned_at": nil}).Error; err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(xerr.NewDBErr(), "unpin others err")
		}
		if err := tx.Table("group_announcements").
			Where("group_id = ? AND id = ? AND deleted = 0", groupIdInt, in.AnnouncementId).
			Updates(map[string]any{"pinned": 1, "pinned_at": now}).Error; err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(xerr.NewDBErr(), "pin err")
		}
	} else {
		if err := tx.Table("group_announcements").
			Where("group_id = ? AND id = ?", groupIdInt, in.AnnouncementId).
			Updates(map[string]any{"pinned": 0, "pinned_at": nil}).Error; err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(xerr.NewDBErr(), "unpin err")
		}
	}

	// 同步 groups.notification：优先 pinned，否则最新一条
	type annRow struct {
		Content   string `gorm:"column:content"`
		CreatedBy string `gorm:"column:created_by"`
	}
	var ann annRow
	err = tx.Table("group_announcements").
		Select("content, created_by").
		Where("group_id = ? AND deleted = 0 AND pinned = 1", groupIdInt).
		Order("pinned_at desc").
		First(&ann).Error
	if err != nil {
		// 没有 pinned，则取最新
		err2 := tx.Table("group_announcements").
			Select("content, created_by").
			Where("group_id = ? AND deleted = 0", groupIdInt).
			Order("created_at desc").
			First(&ann).Error
		if err2 != nil {
			ann.Content = ""
			ann.CreatedBy = ""
		}
	}

	uidInt, _ := strconv.ParseUint(ann.CreatedBy, 10, 64)
	if err := tx.Table("groups").Where("id = ?", groupIdInt).
		Updates(map[string]any{"notification": ann.Content, "notification_uid": uidInt, "updated_at": now}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group notification err")
	}

	tx.Commit()
	return &social.GroupAnnouncementPinResp{}, nil
}
