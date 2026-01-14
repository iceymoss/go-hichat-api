package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupAnnouncementListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupAnnouncementListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAnnouncementListLogic {
	return &GroupAnnouncementListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupAnnouncementListLogic) GroupAnnouncementList(in *social.GroupAnnouncementListReq) (*social.GroupAnnouncementListResp, error) {
	// 必须是群成员
	_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}

	groupIdInt, err := strconv.Atoi(in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}

	type row struct {
		Id        int64      `gorm:"column:id"`
		Content   string     `gorm:"column:content"`
		CreatedBy string     `gorm:"column:created_by"`
		CreatedAt *time.Time `gorm:"column:created_at"`
		Pinned    bool       `gorm:"column:pinned"`
		PinnedAt  *time.Time `gorm:"column:pinned_at"`
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	q := mysqlConn.Table("group_announcements").
		Where("group_id = ?", groupIdInt)
	if !in.IncludeDeleted {
		q = q.Where("deleted = 0")
	}
	// proto request 没有分页字段，这里先固定取前 200 条
	limit := 200

	var rows []row
	if err := q.Order("pinned desc, pinned_at desc, created_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list announcement err")
	}

	list := make([]*social.GroupAnnouncement, 0, len(rows))
	for _, r := range rows {
		createdAt := int64(0)
		if r.CreatedAt != nil {
			createdAt = r.CreatedAt.Unix()
		}
		pinnedAt := int64(0)
		if r.PinnedAt != nil {
			pinnedAt = r.PinnedAt.Unix()
		}
		list = append(list, &social.GroupAnnouncement{
			Id:        r.Id,
			GroupId:   in.GroupId,
			Content:   r.Content,
			CreatedBy: r.CreatedBy,
			CreatedAt: createdAt,
			Pinned:    r.Pinned,
			PinnedAt:  pinnedAt,
		})
	}

	return &social.GroupAnnouncementListResp{List: list}, nil
}
