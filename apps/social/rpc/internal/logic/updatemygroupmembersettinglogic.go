package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
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

// UpdateMyGroupMemberSetting 更新我的群设置，直接写 group_members 表
func (l *UpdateMyGroupMemberSettingLogic) UpdateMyGroupMemberSetting(in *social.UpdateMyGroupMemberSettingReq) (*social.UpdateMyGroupMemberSettingResp, error) {
	// 必须是群成员
	_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("不在群中"), "not in group")
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	updates := map[string]any{}
	if in.GroupNickname != "" {
		updates["group_nickname"] = in.GroupNickname
	}
	if in.GroupRemark != "" {
		updates["group_remark"] = in.GroupRemark
	}
	if len(updates) == 0 {
		return &social.UpdateMyGroupMemberSettingResp{}, nil
	}

	err = mysqlConn.Table("group_members").
		Where("group_id = ? AND user_id = ?", in.GroupId, in.UserId).
		Updates(updates).Error
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update group member setting err")
	}

	return &social.UpdateMyGroupMemberSettingResp{}, nil
}
