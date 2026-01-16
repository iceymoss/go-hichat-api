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

type GetMyGroupMemberSettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyGroupMemberSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyGroupMemberSettingLogic {
	return &GetMyGroupMemberSettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 群成员资料（群内昵称/群备注/@列表）
func (l *GetMyGroupMemberSettingLogic) GetMyGroupMemberSetting(in *social.GetMyGroupMemberSettingReq) (*social.GetMyGroupMemberSettingResp, error) {
	// 必须是群成员
	_, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("no permission"), "not in group")
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	var r struct {
		GroupNickname string `gorm:"column:group_nickname"`
		GroupRemark   string `gorm:"column:group_remark"`
		UpdatedAt     *int64 `gorm:"column:updated_at"`
	}
	// 直接查，不存在返回空
	err = mysqlConn.Table("group_member_settings").
		Select("group_nickname, group_remark, UNIX_TIMESTAMP(updated_at) as updated_at").
		Where("group_id = ? AND user_id = ?", in.GroupId, in.UserId).
		Scan(&r).Error
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "get group member setting err")
	}

	updatedAt := int64(0)
	if r.UpdatedAt != nil {
		updatedAt = *r.UpdatedAt
	}

	return &social.GetMyGroupMemberSettingResp{
		Setting: &social.GroupMemberSetting{
			GroupId:       in.GroupId,
			UserId:        in.UserId,
			GroupNickname: r.GroupNickname,
			GroupRemark:   r.GroupRemark,
			UpdatedAt:     updatedAt,
		},
	}, nil
}
