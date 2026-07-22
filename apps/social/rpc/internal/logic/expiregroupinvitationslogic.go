package logic

import (
	"context"
	"errors"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ExpireGroupInvitationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExpireGroupInvitationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExpireGroupInvitationsLogic {
	return &ExpireGroupInvitationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExpireGroupInvitationsLogic) ExpireGroupInvitations(in *social.ExpireGroupInvitationsReq) (*social.ExpireGroupInvitationsResp, error) {
	if err := rpcauth.RequireTask(l.ctx); err != nil {
		if errors.Is(err, rpcauth.ErrWrongPrincipal) {
			return nil, status.Error(codes.PermissionDenied, "task principal is required")
		}
		return nil, status.Error(codes.Unauthenticated, "missing or invalid rpc caller principal")
	}
	limit := int(in.GetBatchSize())
	if limit == 0 {
		limit = 200
	} else if limit < 1 {
		limit = 1
	} else if limit > 500 {
		limit = 500
	}
	now := time.Now()
	var expired, selected int32
	err := transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
		var invitations []objects.GroupInvitation
		if err := tx.Select("id").Where("status = ? AND expires_at <= ?", groupInvitationPending, now).
			Order("id ASC").Limit(limit).Find(&invitations).Error; err != nil {
			return err
		}
		selected = int32(len(invitations))
		ids := make([]uint64, 0, len(invitations))
		for _, invitation := range invitations {
			result := tx.Model(&objects.GroupInvitation{}).
				Where("id = ? AND status = ? AND expires_at <= ?", invitation.ID, groupInvitationPending, now).
				Updates(map[string]any{"status": groupInvitationExpired, "handled_at": now})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 1 {
				ids = append(ids, invitation.ID)
			}
		}
		if err := resolveInviteReceipts(tx, ids, receiptExpired, now, ""); err != nil {
			return err
		}
		expired = int32(len(ids))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to expire group invitations")
	}

	return &social.ExpireGroupInvitationsResp{Expired: expired, HasMore: selected == int32(limit)}, nil
}
