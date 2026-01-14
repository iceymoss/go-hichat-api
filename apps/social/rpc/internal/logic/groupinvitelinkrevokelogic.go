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

type GroupInviteLinkRevokeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupInviteLinkRevokeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLinkRevokeLogic {
	return &GroupInviteLinkRevokeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupInviteLinkRevokeLogic) GroupInviteLinkRevoke(in *social.GroupInviteLinkRevokeReq) (*social.GroupInviteLinkRevokeResp, error) {
	if in.Token == "" {
		return nil, errors.Wrapf(xerr.NewMsg("token required"), "token required")
	}

	// 至少群成员
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}

	groupIdInt, err := strconv.Atoi(in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}

	type row struct {
		CreatedBy string `gorm:"column:created_by"`
		Revoked   bool   `gorm:"column:revoked"`
	}
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	var r row
	if err := mysqlConn.Table("group_invite_links").
		Select("created_by, revoked").
		Where("group_id = ? AND token = ?", groupIdInt, in.Token).
		First(&r).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("link not found"), "link not found")
	}
	if r.Revoked {
		return &social.GroupInviteLinkRevokeResp{}, nil
	}

	// 权限：管理员/群主 或 链接创建者
	if member.RoleLevel < int(constants.ManagerGroupRoleLevel) && r.CreatedBy != in.UserId {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "no permission to revoke")
	}

	now := time.Now()
	if err := mysqlConn.Table("group_invite_links").
		Where("group_id = ? AND token = ? AND revoked = 0", groupIdInt, in.Token).
		Updates(map[string]any{"revoked": 1, "revoked_at": now}).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "revoke err")
	}

	return &social.GroupInviteLinkRevokeResp{}, nil
}
