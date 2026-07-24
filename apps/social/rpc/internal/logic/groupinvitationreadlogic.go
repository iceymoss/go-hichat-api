package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GroupInvitationReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupInvitationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationReadLogic {
	return &GroupInvitationReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupInvitationReadLogic) GroupInvitationRead(in *social.GroupInvitationReadReq) (*social.GroupInvitationReadResp, error) {
	if in.ActorUid == "" {
		return nil, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if _, err := parsePositiveID(in.ActorUid, "actor uid"); err != nil {
		return nil, err
	}
	if len(in.InvitationIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invitation ids are required")
	}
	var counts receiptCounts
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := markReceiptsRead(tx, in.ActorUid, receiptTypeGroupInvite, in.InvitationIds); err != nil {
			return err
		}
		var err error
		counts, err = countUnreadReceipts(tx, in.ActorUid, true)
		return err
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark invitation receipts read")
	}
	return &social.GroupInvitationReadResp{Count: counts.Total, Apply: counts.Apply, Result: counts.Result, Invite: counts.Invite}, nil
}
