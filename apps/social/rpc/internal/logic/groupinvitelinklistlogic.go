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

type GroupInviteLinkListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupInviteLinkListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLinkListLogic {
	return &GroupInviteLinkListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupInviteLinkListLogic) GroupInviteLinkList(in *social.GroupInviteLinkListReq) (*social.GroupInviteLinkListResp, error) {
	// 权限：至少管理员/群主可查看链接列表
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}
	if member.RoleLevel < int(constants.ManagerGroupRoleLevel) {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only admin/owner can list invite links")
	}

	groupIdInt, err := strconv.Atoi(in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}

	type row struct {
		Token     string     `gorm:"column:token"`
		GroupId   int        `gorm:"column:group_id"`
		CreatedBy string     `gorm:"column:created_by"`
		ExpireAt  *time.Time `gorm:"column:expire_at"`
		MaxUses   int32      `gorm:"column:max_uses"`
		UsedCount int32      `gorm:"column:used_count"`
		Revoked   bool       `gorm:"column:revoked"`
		RevokedAt *time.Time `gorm:"column:revoked_at"`
		CreatedAt *time.Time `gorm:"column:created_at"`
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	q := mysqlConn.Table("group_invite_links").Where("group_id = ?", groupIdInt)
	if !in.IncludeRevoked {
		q = q.Where("revoked = 0")
	}
	q = q.Order("created_at desc")

	var rows []row
	if err := q.Find(&rows).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "list invite links err")
	}

	list := make([]*social.GroupInviteLink, 0, len(rows))
	for _, r := range rows {
		exp := int64(0)
		if r.ExpireAt != nil {
			exp = r.ExpireAt.Unix()
		}
		revAt := int64(0)
		if r.RevokedAt != nil {
			revAt = r.RevokedAt.Unix()
		}
		createdAt := int64(0)
		if r.CreatedAt != nil {
			createdAt = r.CreatedAt.Unix()
		}
		list = append(list, &social.GroupInviteLink{
			Token:     r.Token,
			GroupId:   in.GroupId,
			CreatedBy: r.CreatedBy,
			CreatedAt: createdAt,
			ExpireAt:  exp,
			MaxUses:   r.MaxUses,
			UsedCount: r.UsedCount,
			Revoked:   r.Revoked,
			RevokedAt: revAt,
		})
	}

	return &social.GroupInviteLinkListResp{List: list}, nil
}
