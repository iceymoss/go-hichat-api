package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FriendPutInLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewFriendPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInLogic {
	return &FriendPutInLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

// FriendPutIn creates one active request per directed user pair.
func (l *FriendPutInLogic) FriendPutIn(in *social.FriendPutInReq) (*social.FriendPutInResp, error) {
	actor, target, err := validateFriendActors(in.ActorUid, in.UserId, in.ReqUid)
	if err != nil {
		return nil, err
	}
	if err := l.requireNormalUsers(in.ActorUid, in.ReqUid); err != nil {
		return nil, err
	}

	var relationCount int64
	err = l.DB.WithContext(l.ctx).Model(&objects.Friend{}).
		Where("(user_id = ? AND friend_uid = ?) OR (user_id = ? AND friend_uid = ?)", actor, target, target, actor).
		Count(&relationCount).Error
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to inspect friend relation")
	}
	switch relationCount {
	case 2:
		return &social.FriendPutInResp{Status: 1, AlreadyFriend: true}, nil
	case 1:
		return nil, status.Error(codes.FailedPrecondition, "one-way friend relation conflict")
	}

	activeKey := fmt.Sprintf("friend:%d:%d", actor, target)
	pending := 0
	visible := 1
	reqTime := time.Unix(in.ReqTime, 0).In(utils.ChinaLocation)
	if in.ReqTime <= 0 {
		reqTime = utils.NowInChina()
	}
	request := &objects.FriendRequest{
		UserID: actor, ReqUID: target, ReqMsg: in.ReqMsg, ReqTime: reqTime,
		HandleResult: &pending, Status: &visible, Remark: in.Remark, ActiveKey: &activeKey,
	}
	created := false
	err = transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		created = false
		request.ID = 0
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(request)
		if result.Error != nil {
			return result.Error
		}
		created = result.RowsAffected == 1
		if !created {
			return tx.Where("active_key = ?", activeKey).First(request).Error
		}
		if err := createReceipt(tx, receiptTypeFriend, request.ID, in.ReqUid, receiptKindApply, 1, receiptPending, reqTime, false, nil); err != nil {
			return err
		}
		return emitRequestNotificationInTx(tx, l.ServiceContext, NotifyFriendApply, receiptTypeFriend, request.ID, in.ReqUid, in.ActorUid, 0, receiptPending, in.ReqMsg)
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create friend request")
	}

	if created {
		if l.FriendRequestsModel != nil {
			l.FriendRequestsModel.InvalidateCountCache(l.ctx, in.ReqUid)
		}
	}
	return &social.FriendPutInResp{RequestId: int64(request.ID), Status: 0, AlreadyPending: !created}, nil
}

func validateFriendActors(actorUID, legacyUID, targetUID string) (uint64, uint64, error) {
	if actorUID == "" {
		return 0, 0, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if legacyUID != "" && legacyUID != actorUID {
		return 0, 0, status.Error(codes.PermissionDenied, "actor uid does not match user id")
	}
	actor, err := strconv.ParseUint(actorUID, 10, 64)
	if err != nil || actor == 0 {
		return 0, 0, status.Error(codes.InvalidArgument, "actor uid must be a positive integer")
	}
	target, err := strconv.ParseUint(targetUID, 10, 64)
	if err != nil || target == 0 {
		return 0, 0, status.Error(codes.InvalidArgument, "target uid must be a positive integer")
	}
	if actor == target {
		return 0, 0, status.Error(codes.InvalidArgument, "cannot send a friend request to yourself")
	}
	return actor, target, nil
}

func (l *FriendPutInLogic) requireNormalUsers(ids ...string) error {
	for _, id := range ids {
		resp, err := l.User.GetUserById(l.ctx, &user.GetUserByIdRequest{Id: id})
		if err != nil {
			return normalizeUserLookupError(err)
		}
		if resp == nil || resp.User == nil || resp.User.Status != 1 {
			return status.Error(codes.NotFound, "user does not exist or is unavailable")
		}
	}
	return nil
}

func normalizeUserLookupError(err error) error {
	switch status.Code(err) {
	case codes.NotFound:
		return status.Error(codes.NotFound, "user does not exist or is unavailable")
	case codes.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "user service deadline exceeded")
	case codes.Unavailable:
		return status.Error(codes.Unavailable, "user service unavailable")
	default:
		return status.Error(codes.Internal, "failed to validate user")
	}
}
