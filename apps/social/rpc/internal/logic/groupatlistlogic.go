package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GroupAtListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupAtListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAtListLogic {
	return &GroupAtListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupAtListLogic) GroupAtList(in *social.GroupAtListReq) (*social.GroupAtListResp, error) {
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
		MemberUid     string `gorm:"column:user_id"`
		RoleLevel     int32  `gorm:"column:role_level"`
		GroupNickname string `gorm:"column:group_nickname"`
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	q := mysqlConn.Table("group_members gm").
		Select("CAST(gm.user_id AS CHAR) as user_id, gm.role_level, COALESCE(gms.group_nickname,'') as group_nickname").
		Joins("LEFT JOIN group_member_settings gms ON gms.group_id = gm.group_id AND gms.user_id = gm.user_id").
		Where("gm.group_id = ?", groupIdInt)

	if in.Keyword != "" {
		// 仅基于群内昵称过滤（用户昵称/拼音排序在 social-api 层可用 user-rpc 丰富）
		q = q.Where("gms.group_nickname LIKE ? OR CAST(gm.user_id AS CHAR) = ?", "%"+in.Keyword+"%", in.Keyword)
	}

	var rows []row
	// proto request 没有分页字段，这里先固定取前 200 条
	limit := 200
	if err := q.Order("gm.role_level desc, gms.group_nickname asc, gm.user_id asc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "group at list err")
	}

	out := make([]*social.AtMember, 0, len(rows))
	for _, r := range rows {
		out = append(out, &social.AtMember{
			UserId:        r.MemberUid,
			RoleLevel:     r.RoleLevel,
			GroupNickname: r.GroupNickname,
			// nickname/avatar 由 social-api 调用 user-rpc 补齐（避免 social-rpc 依赖 user-rpc）
			Nickname: "",
			Avatar:   "",
		})
	}
	return &social.GroupAtListResp{List: out}, nil
}
