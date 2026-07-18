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
	if in.FriendReqId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "friend request id must be positive")
	}
	if _, err := strconv.ParseUint(in.UserId, 10, 64); err != nil || in.UserId == "0" {
		return nil, status.Error(codes.InvalidArgument, "user id must be a positive integer")
	}
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		var request objects.FriendRequest
		if err := tx.First(&request, uint64(in.FriendReqId)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "friend request not found")
			}
			return err
		}
		if strconv.FormatUint(request.UserID, 10) != in.UserId && strconv.FormatUint(request.ReqUID, 10) != in.UserId {
			return status.Error(codes.PermissionDenied, "friend request does not belong to current user")
		}
		now := time.Now()
		return tx.Model(&objects.SocialRequestReceipt{}).
			Where("request_type = ? AND request_id = ? AND receiver_id = ?", receiptTypeFriend, request.ID, in.UserId).
			Updates(map[string]any{"is_read": 1, "is_actionable": 0, "read_at": now}).Error
	})
	if err != nil {
		if status.Code(err) != codes.Unknown {
			return nil, err
		}
		return nil, status.Error(codes.Internal, "failed to hide friend request")
	}

	return &social.FriendPutInDeleteResp{}, nil
}
