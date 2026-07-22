package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
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
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
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

	var putinResp *social.GroupPutinResp
	err = transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
		putinResp = nil
		var link linkRow
		if err := tx.Table("group_invite_links").Clauses(clause.Locking{Strength: "UPDATE"}).Where("token = ?", in.Token).First(&link).Error; err != nil {
			return errors.Wrapf(xerr.NewMsg("invalid token"), "invalid token")
		}
		if err := requireNormalUser(l.ctx, l.svcCtx.User, actor); err != nil {
			return err
		}
		if link.Revoked {
			return errors.Wrapf(xerr.NewMsg("token revoked"), "token revoked")
		}
		if link.ExpireAt != nil && link.ExpireAt.Before(time.Now()) {
			return errors.Wrapf(xerr.NewMsg("token expired"), "token expired")
		}
		if link.MaxUses > 0 && link.UsedCount >= link.MaxUses {
			return errors.Wrapf(xerr.NewMsg("token used up"), "token used up")
		}

		groupID := strconv.Itoa(link.GroupId)
		putin := NewGroupPutinLogic(l.ctx, l.svcCtx)
		var err error
		putinResp, err = putin.groupPutInByTokenTx(tx, &social.GroupPutinReq{
			GroupId: groupID, ReqId: actor, ReqMsg: in.ReqMsg,
			JoinSource: int32(constants.InviteLinkGroupJoinSource), InviterUid: link.CreatedBy, ActorUid: actor,
		})
		if err != nil {
			return err
		}
		if putinResp.AlreadyMember || link.MaxUses == 0 {
			return nil
		}
		result := tx.Table("group_invite_links").Where("token = ? AND revoked = ? AND used_count = ?", in.Token, false, link.UsedCount).
			UpdateColumn("used_count", gorm.Expr("used_count + ?", 1))
		if result.Error != nil {
			return errors.Wrapf(xerr.NewDBErr(), "update used_count err")
		}
		if result.RowsAffected != 1 {
			return errors.Wrapf(xerr.NewMsg("token used up"), "token usage changed concurrently")
		}
		return nil
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to join group by token")
	}
	return &social.GroupJoinByTokenResp{GroupId: putinResp.GroupId, GroupIdString: putinResp.GroupIdString, IsPass: putinResp.IsPass}, nil
}
