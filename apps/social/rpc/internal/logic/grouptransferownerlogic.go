package logic

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupTransferOwnerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupTransferOwnerLogic {
	return &GroupTransferOwnerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupTransferOwnerLogic) GroupTransferOwner(in *social.GroupTransferOwnerReq) (*social.GroupTransferOwnerResp, error) {
	// 1) 校验群存在且操作者为群主
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group not found %v", in.GroupId)
	}
	if group.Status == 1 {
		return nil, errors.Wrapf(xerr.NewMsg("group disbanded"), "group disbanded")
	}
	if group.CreatorUid != in.UserId {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "only owner can transfer ownership")
	}
	if in.NewOwnerId == "" || in.NewOwnerId == in.UserId {
		return nil, errors.Wrapf(xerr.NewMsg("invalid new owner"), "invalid new owner")
	}

	// 2) 新群主必须已在群内
	_, err = l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.NewOwnerId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("new owner not in group"), "new owner not in group")
	}

	now := time.Now()
	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()

	// 3) 更新 groups.creator_uid
	if err := tx.Table("groups").Where("id = ?", group.Id).Updates(map[string]any{
		"creator_uid": in.NewOwnerId,
		"updated_at":  now,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group owner err")
	}

	// 4) 更新角色：新群主 -> 2，旧群主 -> 1/0
	if err := tx.Table("group_members").Where("group_id = ? AND user_id = ?", in.GroupId, in.NewOwnerId).
		Updates(map[string]any{"role_level": int(constants.CreatorGroupRoleLevel)}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "set new owner role err")
	}

	oldRole := int(constants.AtLargeGroupRoleLevel)
	// 默认保持管理员（微信类似体验）
	if in.KeepOldOwnerAsAdmin {
		oldRole = int(constants.ManagerGroupRoleLevel)
	}
	if err := tx.Table("group_members").Where("group_id = ? AND user_id = ?", in.GroupId, in.UserId).
		Updates(map[string]any{"role_level": oldRole}).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewDBErr(), "set old owner role err")
	}

	tx.Commit()

	return &social.GroupTransferOwnerResp{}, nil
}
