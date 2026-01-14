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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GroupJoinByTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupJoinByTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupJoinByTokenLogic {
	return &GroupJoinByTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupJoinByTokenLogic) GroupJoinByToken(in *social.GroupJoinByTokenReq) (*social.GroupJoinByTokenResp, error) {
	if in.Token == "" {
		return nil, errors.Wrapf(xerr.NewMsg("token required"), "token required")
	}

	type linkRow struct {
		GroupId   int        `gorm:"column:group_id"`
		CreatedBy string     `gorm:"column:created_by"`
		ExpireAt  *time.Time `gorm:"column:expire_at"`
		MaxUses   int32      `gorm:"column:max_uses"`
		UsedCount int32      `gorm:"column:used_count"`
		Revoked   bool       `gorm:"column:revoked"`
		RevokedAt *time.Time `gorm:"column:revoked_at"`
	}

	mysqlConn := db.GetMysqlConn(db.MYSQL_DB_HICHAT2)
	tx := mysqlConn.Begin()

	// 行锁，避免并发超用
	var link linkRow
	if err := tx.Table("group_invite_links").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token = ?", in.Token).
		First(&link).Error; err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewMsg("invalid token"), "invalid token")
	}

	if link.Revoked {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewMsg("token revoked"), "token revoked")
	}
	if link.ExpireAt != nil && link.ExpireAt.Before(time.Now()) {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewMsg("token expired"), "token expired")
	}
	if link.MaxUses > 0 && link.UsedCount >= link.MaxUses {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewMsg("token used up"), "token used up")
	}

	groupIdStr := strconv.Itoa(link.GroupId)
	group, err := l.svcCtx.GroupsModel.FindOne(l.ctx, groupIdStr)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewMsg("group not found"), "group not found")
	}
	if group.Status == 1 {
		tx.Rollback()
		return nil, errors.Wrapf(xerr.NewMsg("group disbanded"), "group disbanded")
	}

	// 已在群内：不消耗次数，直接返回
	if _, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, groupIdStr); err == nil {
		tx.Rollback()
		return &social.GroupJoinByTokenResp{GroupId: int32(link.GroupId), IsPass: 1}, nil
	}

	// 统一走 GroupPutin 逻辑（joinSource=InviteLink），由群配置/邀请人角色决定是否直入群
	putin := NewGroupPutinLogic(l.ctx, l.svcCtx)
	putinResp, err := putin.GroupPutin(&social.GroupPutinReq{
		GroupId:    groupIdStr,
		ReqId:      in.UserId,
		ReqMsg:     in.ReqMsg,
		ReqTime:    time.Now().Unix(),
		JoinSource: int32(constants.InviteLinkGroupJoinSource),
		InviterUid: link.CreatedBy, // 链接创建人作为 inviter（用于“管理员/群主邀请可直入群”规则）
	})
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 消耗一次使用次数（仅当 max_uses > 0）
	if link.MaxUses > 0 {
		if err := tx.Table("group_invite_links").
			Where("token = ? AND revoked = 0", in.Token).
			UpdateColumn("used_count", gorm.Expr("used_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return nil, errors.Wrapf(xerr.NewDBErr(), "update used_count err")
		}
	}

	tx.Commit()
	return &social.GroupJoinByTokenResp{GroupId: putinResp.GroupId, IsPass: putinResp.IsPass}, nil
}
