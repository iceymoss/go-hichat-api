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
	"gorm.io/gorm/clause"
)

type UpdateMyGroupMemberSettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMyGroupMemberSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMyGroupMemberSettingLogic {
	return &UpdateMyGroupMemberSettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateMyGroupMemberSettingLogic) UpdateMyGroupMemberSetting(in *social.UpdateMyGroupMemberSettingReq) (*social.UpdateMyGroupMemberSettingResp, error) {
	// 必须是群成员
	_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	now := time.Now()
	uidInt, err := strconv.ParseUint(in.UserId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid userId"), "invalid userId")
	}
	gidInt, err := strconv.ParseUint(in.GroupId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("invalid groupId"), "invalid groupId")
	}
	// upsert
	err = mysqlConn.Table("group_member_settings").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"group_nickname", "group_remark", "updated_at"}),
	}).Create(map[string]any{
		"group_id":       gidInt,
		"user_id":        uidInt,
		"group_nickname": in.GroupNickname,
		"group_remark":   in.GroupRemark,
		"updated_at":     now,
	}).Error
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group member setting err")
	}

	return &social.UpdateMyGroupMemberSettingResp{}, nil
}
