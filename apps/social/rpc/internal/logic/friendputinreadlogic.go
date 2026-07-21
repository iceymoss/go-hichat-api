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

type FriendPutInReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInReadLogic {
	return &FriendPutInReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendPutInReadLogic) FriendPutInRead(in *social.FriendPutInReadReq) (*social.FriendPutInReadResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	ids := append([]uint64(nil), in.RequestIds...)
	if in.FriendReqId > 0 {
		ids = append(ids, uint64(in.FriendReqId))
	}
	var counts receiptCounts
	err = transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
		if err := markReceiptsRead(tx, actor, receiptTypeFriend, ids); err != nil {
			return err
		}
		var err error
		counts, err = countUnreadReceipts(tx, actor, false)
		return err
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark friend request receipts read")
	}
	return &social.FriendPutInReadResp{Count: counts.Total, Apply: counts.Apply, Result: counts.Result}, nil
}
