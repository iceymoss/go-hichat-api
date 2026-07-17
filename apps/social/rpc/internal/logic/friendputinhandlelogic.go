package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/utils"

	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errFriendRequestCASMiss = errors.New("friend request CAS missed")

type FriendPutInHandleLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewFriendPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInHandleLogic {
	return &FriendPutInHandleLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

// FriendPutInHandle changes a pending request exactly once.
func (l *FriendPutInHandleLogic) FriendPutInHandle(in *social.FriendPutInHandleReq) (*social.FriendPutInHandleResp, error) {
	if in.FriendReqId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "friend request id must be positive")
	}
	if in.HandleResult != 1 && in.HandleResult != 2 {
		return nil, status.Error(codes.InvalidArgument, "handle result must be 1 or 2")
	}
	actor, err := validateHandleActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}

	var request objects.FriendRequest
	if err := l.DB.WithContext(l.ctx).First(&request, uint64(in.FriendReqId)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "friend request not found")
		}
		return nil, status.Error(codes.Internal, "failed to load friend request")
	}
	if request.ReqUID != actor {
		return nil, status.Error(codes.PermissionDenied, "only the requested user may handle this request")
	}
	if request.UserID == 0 || request.ReqUID == 0 || request.UserID == request.ReqUID {
		return nil, status.Error(codes.FailedPrecondition, "friend request has invalid participants")
	}

	actorUser, err := l.normalUser(strconv.FormatUint(actor, 10))
	if err != nil {
		return nil, err
	}
	applicantUser, err := l.normalUser(strconv.FormatUint(request.UserID, 10))
	if err != nil {
		return nil, err
	}

	tagsJSON := ""
	if in.HandleResult == 1 && len(in.Tags) > 0 {
		data, err := jsonx.Marshal(in.Tags)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid friend tags")
		}
		tagsJSON = string(data)
	}

	changed := false
	var relationEvent *mq.RelationChangeTransfer
	err = l.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		now := utils.NowInChina()
		result := tx.Model(&objects.FriendRequest{}).
			Where("id = ? AND req_uid = ? AND handle_result = ? AND status = ?", request.ID, actor, 0, 1).
			Updates(map[string]any{
				"handle_result": in.HandleResult, "handle_msg": in.HandleMsg, "handled_at": now,
				"active_key": nil, "sender_read": 0,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errFriendRequestCASMiss
		}
		changed = true

		if in.HandleResult == 2 {
			return nil
		}
		var relationCount int64
		if err := tx.Model(&objects.Friend{}).
			Where("(user_id = ? AND friend_uid = ?) OR (user_id = ? AND friend_uid = ?)", request.UserID, request.ReqUID, request.ReqUID, request.UserID).
			Count(&relationCount).Error; err != nil {
			return err
		}
		if relationCount == 1 {
			return status.Error(codes.FailedPrecondition, "one-way friend relation conflict")
		}

		applicantRemark := request.Remark
		if applicantRemark == "" {
			applicantRemark = actorUser.Nickname
		}
		actorRemark := in.Remark
		if actorRemark == "" {
			actorRemark = applicantUser.Nickname
		}
		createdAt := time.Now()
		friends := []objects.Friend{
			{UserID: request.UserID, FriendUID: request.ReqUID, Remark: applicantRemark, AddSource: intPtr(1), CreatedAt: &createdAt},
			{UserID: request.ReqUID, FriendUID: request.UserID, Remark: actorRemark, AddSource: intPtr(1), FriendTags: tagsJSON, CreatedAt: &createdAt},
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&friends).Error; err != nil {
			return err
		}
		if err := tx.Model(&objects.Friend{}).
			Where("(user_id = ? AND friend_uid = ?) OR (user_id = ? AND friend_uid = ?)", request.UserID, request.ReqUID, request.ReqUID, request.UserID).
			Count(&relationCount).Error; err != nil || relationCount != 2 {
			if err != nil {
				return err
			}
			return status.Error(codes.Internal, "failed to establish bidirectional friendship")
		}

		if err := tx.Model(&objects.FriendRequest{}).
			Where("id <> ? AND user_id = ? AND req_uid = ? AND handle_result = ? AND status = ?", request.ID, request.ReqUID, request.UserID, 0, 1).
			Updates(map[string]any{"handle_result": 1, "handled_at": now, "active_key": nil, "sender_read": 1}).Error; err != nil {
			return err
		}
		relationEvent = &mq.RelationChangeTransfer{
			FriendA: strconv.FormatUint(request.UserID, 10), FriendB: strconv.FormatUint(request.ReqUID, 10),
			OperatorId: in.ActorUid,
		}
		return emitRelationChangeInTx(tx, l.ServiceContext, constants.RelationEventFriendAdded, "", relationEvent)
	})
	if err != nil {
		if errors.Is(err, errFriendRequestCASMiss) {
			var latest objects.FriendRequest
			if loadErr := l.DB.WithContext(l.ctx).First(&latest, request.ID).Error; loadErr != nil {
				return nil, status.Error(codes.Internal, "failed to reload friend request")
			}
			if latest.ReqUID != actor {
				return nil, status.Error(codes.PermissionDenied, "only the requested user may handle this request")
			}
			if latest.HandleResult != nil && *latest.HandleResult == int(in.HandleResult) {
				return &social.FriendPutInHandleResp{RequestId: int64(request.ID), HandleResult: in.HandleResult, Idempotent: true}, nil
			}
			return nil, status.Error(codes.FailedPrecondition, "friend request already handled with a different result")
		}
		if status.Code(err) != codes.Unknown {
			return nil, err
		}
		return nil, status.Error(codes.Internal, "failed to handle friend request")
	}
	if !changed {
		return &social.FriendPutInHandleResp{RequestId: int64(request.ID), HandleResult: in.HandleResult, Idempotent: true}, nil
	}

	if relationEvent != nil {
		bestEffortApplyCache(l.ctx, l.RelationCache, relationEvent, 0)
	}
	if l.FriendRequestsModel != nil {
		l.FriendRequestsModel.InvalidateCountCache(l.ctx, strconv.FormatUint(request.UserID, 10), strconv.FormatUint(request.ReqUID, 10))
	}
	notifyType := NotifyFriendReject
	if in.HandleResult == 1 {
		notifyType = NotifyFriendAccept
	}
	emitCommonNotify(l.ctx, l.ServiceContext, &mq.CommonNotify{
		NotifyType: notifyType, ReceiverId: strconv.FormatUint(request.UserID, 10), ActorId: in.ActorUid,
		BizId: fmt.Sprintf("friend.handle:%d:%d", request.ID, in.HandleResult), Content: in.HandleMsg,
	})
	return &social.FriendPutInHandleResp{RequestId: int64(request.ID), HandleResult: in.HandleResult}, nil
}

func validateHandleActor(actorUID, legacyUID string) (uint64, error) {
	if actorUID == "" {
		return 0, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if legacyUID != "" && legacyUID != actorUID {
		return 0, status.Error(codes.PermissionDenied, "actor uid does not match user id")
	}
	actor, err := strconv.ParseUint(actorUID, 10, 64)
	if err != nil || actor == 0 {
		return 0, status.Error(codes.InvalidArgument, "actor uid must be a positive integer")
	}
	return actor, nil
}

func (l *FriendPutInHandleLogic) normalUser(id string) (*user.UserEntity, error) {
	resp, err := l.User.GetUserById(l.ctx, &user.GetUserByIdRequest{Id: id})
	if err != nil {
		return nil, normalizeUserLookupError(err)
	}
	if resp == nil || resp.User == nil || resp.User.Status != 1 {
		return nil, status.Error(codes.NotFound, "user does not exist or is unavailable")
	}
	return resp.User, nil
}

func intPtr(v int) *int { return &v }
