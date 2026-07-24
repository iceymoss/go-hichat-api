package logic

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GroupInvitationListLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupInvitationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationListLogic {
	return &GroupInvitationListLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupInvitationListLogic) GroupInvitationList(in *social.GroupInvitationListReq) (*social.GroupInvitationListResp, error) {
	if in.ActorUid == "" {
		return nil, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	actor, err := parsePositiveID(in.ActorUid, "actor uid")
	if err != nil {
		return nil, err
	}
	if in.Status < -1 || in.Status > groupInvitationExpired {
		return nil, status.Error(codes.InvalidArgument, "invitation status must be between 0 and 4, or -1")
	}
	page, size := in.Page, in.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		return nil, status.Error(codes.InvalidArgument, "page size must not exceed 100")
	}
	if err := l.expireInvitations(actor); err != nil {
		return nil, err
	}

	query := l.DB.WithContext(l.ctx).Model(&objects.GroupInvitation{}).Where("invitee_uid = ?", actor)
	if in.Status >= 0 {
		query = query.Where("status = ?", in.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to count group invitations")
	}
	var rows []objects.GroupInvitation
	if err := query.Order("created_at DESC, id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&rows).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to list group invitations")
	}
	ids := make([]uint64, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	var receipts []objects.SocialRequestReceipt
	if len(ids) > 0 {
		if err := l.DB.WithContext(l.ctx).Where("request_type = ? AND request_id IN ? AND receiver_id = ? AND receipt_kind = ?", receiptTypeGroupInvite, ids, in.ActorUid, receiptKindInvite).Find(&receipts).Error; err != nil {
			return nil, status.Error(codes.Internal, "failed to load invitation receipts")
		}
	}
	byID := make(map[uint64]objects.SocialRequestReceipt, len(receipts))
	for _, receipt := range receipts {
		byID[receipt.RequestID] = receipt
	}
	list := make([]*social.GroupInvitation, len(rows))
	for i := range rows {
		list[i] = invitationProto(&rows[i])
		if receipt, ok := byID[rows[i].ID]; ok {
			list[i].ReadState = int32(receipt.IsRead)
			list[i].Actionable = receipt.IsActionable == 1
		}
	}
	return &social.GroupInvitationListResp{List: list, Total: total}, nil
}

func (l *GroupInvitationListLogic) expireInvitations(actor uint64) error {
	now := time.Now()
	err := transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		var rows []objects.GroupInvitation
		if err := tx.Where("invitee_uid = ? AND status = ? AND expires_at <= ?", actor, groupInvitationPending, now).Find(&rows).Error; err != nil {
			return err
		}
		ids := make([]uint64, 0, len(rows))
		for _, invitation := range rows {
			result := tx.Model(&objects.GroupInvitation{}).Where("id = ? AND status = ? AND expires_at <= ?", invitation.ID, groupInvitationPending, now).
				Updates(map[string]any{"status": groupInvitationExpired, "handled_at": now})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 1 {
				ids = append(ids, invitation.ID)
			}
		}
		return expireInviteReceipts(tx, ids, now)
	})
	if err != nil {
		return status.Error(codes.Internal, "failed to expire group invitations")
	}
	return nil
}
