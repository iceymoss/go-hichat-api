package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FriendPutInDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInDeleteLogic {
	return &FriendPutInDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInDelete hides no shared history; it only closes the caller's personal receipt.
func (l *FriendPutInDeleteLogic) FriendPutInDelete(in *social.FriendPutInDeleteReq) (*social.FriendPutInDeleteResp, error) {
	requestID := in.RequestId
	if requestID == 0 && in.FriendReqId > 0 {
		requestID = uint64(in.FriendReqId)
	}
	if requestID == 0 {
		return nil, status.Error(codes.InvalidArgument, "friend request id must be positive")
	}
	if _, err := strconv.ParseUint(in.UserId, 10, 64); err != nil || in.UserId == "0" {
		return nil, status.Error(codes.InvalidArgument, "user id must be a positive integer")
	}
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var request objects.FriendRequest
		if err := tx.First(&request, requestID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "friend request not found")
			}
			return err
		}
		if strconv.FormatUint(request.UserID, 10) != in.UserId && strconv.FormatUint(request.ReqUID, 10) != in.UserId {
			return status.Error(codes.PermissionDenied, "friend request does not belong to current user")
		}
		now := time.Now()
		kind := receiptKindApply
		if strconv.FormatUint(request.UserID, 10) == in.UserId {
			kind = receiptKindResult
		}
		receipt := objects.SocialRequestReceipt{
			RequestType: receiptTypeFriend, RequestID: request.ID, ReceiverID: in.UserId,
			ReceiptKind: kind, IsRead: 1, IsActionable: 0, Result: receiptInvalidated,
			ReadAt: &now, ResolvedAt: &now, CreatedAt: request.ReqTime,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "request_type"}, {Name: "request_id"}, {Name: "receiver_id"}, {Name: "receipt_kind"}},
			DoUpdates: clause.Assignments(map[string]any{
				"is_read": 1, "is_actionable": 0, "result": receiptInvalidated, "read_at": now, "resolved_at": now,
			}),
		}).Create(&receipt).Error
	})
	if err != nil {
		if status.Code(err) != codes.Unknown {
			return nil, err
		}
		return nil, status.Error(codes.Internal, "failed to hide friend request")
	}

	return &social.FriendPutInDeleteResp{}, nil
}
