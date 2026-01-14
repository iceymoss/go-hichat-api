package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupSetAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupSetAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSetAdminLogic {
	return &GroupSetAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupSetAdminLogic) GroupSetAdmin(in *social.GroupSetAdminReq) (*social.GroupSetAdminResp, error) {
	// 1) 校验群存在且操作者为群主
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found %v", in.GroupId)
	}
	if group.Status == 1 {
		return nil, errors.Wrapf(xerr.NewMsg("group disbanded"), "group disbanded")
	}
	if group.CreatorUid != in.UserId {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only owner can set admin")
	}
	if len(in.MemberIds) == 0 {
		return &social.GroupSetAdminResp{}, nil
	}

	role := int(constants.AtLargeGroupRoleLevel)
	if in.IsAdmin {
		role = int(constants.ManagerGroupRoleLevel)
	}

	// 不允许操作群主本人
	filtered := make([]string, 0, len(in.MemberIds))
	for _, id := range in.MemberIds {
		if id == "" || id == group.CreatorUid {
			continue
		}
		filtered = append(filtered, id)
	}
	if len(filtered) == 0 {
		return &social.GroupSetAdminResp{}, nil
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()

	// 2) 只更新这些成员，且不动 role_level=2（防御）
	if err := tx.Table("group_members").
		Where("group_id = ?", in.GroupId).
		Where("user_id in (?)", filtered).
		Where("role_level <> ?", int(constants.CreatorGroupRoleLevel)).
		Updates(map[string]any{
			"role_level":   role,
			"operator_uid": in.UserId,
		}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "update member role err")
	}

	tx.Commit()

	return &social.GroupSetAdminResp{}, nil
}
