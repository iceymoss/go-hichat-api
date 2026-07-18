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
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var errGroupRequestCASMiss = errors.New("group request CAS missed")

type GroupPutInHandleLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInHandleLogic {
	return &GroupPutInHandleLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupPutInHandleLogic) GroupPutInHandle(in *social.GroupPutInHandleReq) (*social.GroupPutInHandleResp, error) {
	requestID, err := groupHandleRequestID(in)
	if err != nil {
		return nil, err
	}
	if requestID == 0 {
		return nil, status.Error(codes.InvalidArgument, "group request id must be positive")
	}
	if in.HandleResult != 1 && in.HandleResult != 2 {
		return nil, status.Error(codes.InvalidArgument, "handle result must be 1 or 2")
	}
	actor, err := validateGroupHandleActor(in.ActorUid, in.HandleUid)
	if err != nil {
		return nil, err
	}

	var request objects.GroupRequest
	if err := l.DB.WithContext(l.ctx).First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "group request not found")
		}
		return nil, status.Error(codes.Internal, "failed to load group request")
	}
	applicant, err := parsePositiveID(request.ReqID, "request applicant uid")
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "group request has invalid applicant")
	}
	if request.HandleResult != nil && *request.HandleResult != groupRequestPending {
		if err := l.requireCurrentGroupApprover(request.GroupID, actor); err != nil {
			return nil, err
		}
		if *request.HandleResult == int(in.HandleResult) {
			return groupHandleResponse(&request, true), nil
		}
		return nil, status.Error(codes.FailedPrecondition, "group request already handled with a different result")
	}

	changed := false
	err = transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		changed = false
		if _, err := loadNormalGroup(tx, request.GroupID); err != nil {
			return err
		}
		member, err := loadGroupMemberLocked(tx, request.GroupID, actor)
		if err != nil {
			return err
		}
		if member == nil || (member.RoleLevel != 1 && member.RoleLevel != 2) {
			return status.Error(codes.PermissionDenied, "only a current group owner or administrator may handle requests")
		}

		now := time.Now()
		updates := map[string]any{
			"handle_result": in.HandleResult, "handle_user_id": actor, "handle_time": now,
			"active_key": nil,
		}
		if in.HandleResult == groupRequestAccepted {
			actualSource := 1
			if request.SourceType == 2 {
				actualSource = 2
			}
			updates["actual_join_source"] = actualSource
		}
		result := tx.Model(&objects.GroupRequest{}).
			Where("id = ? AND handle_result = ?", request.ID, groupRequestPending).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errGroupRequestCASMiss
		}
		changed = true
		if in.HandleResult == groupRequestRejected {
			return nil
		}

		actualSource := 1
		if request.SourceType == 2 {
			actualSource = 2
		}
		if _, err := createGroupMemberAndOutbox(tx, l.ServiceContext, request.GroupID, applicant, 0, actualSource, request.InviterUserID, actor); err != nil {
			return err
		}
		if err := resolvePendingGroupRequests(tx, request.GroupID, applicant, now, actualSource); err != nil {
			return err
		}
		return tx.Model(&objects.GroupInvitation{}).
			Where("group_id = ? AND invitee_uid = ? AND status = ?", request.GroupID, applicant, groupInvitationPending).
			Updates(map[string]any{"status": groupInvitationInvalidated, "handled_at": now}).Error
	})
	if err != nil {
		if errors.Is(err, errGroupRequestCASMiss) {
			var latest objects.GroupRequest
			if loadErr := l.DB.WithContext(l.ctx).First(&latest, request.ID).Error; loadErr != nil {
				return nil, status.Error(codes.Internal, "failed to reload group request")
			}
			if latest.HandleResult != nil && *latest.HandleResult == int(in.HandleResult) {
				return groupHandleResponse(&latest, true), nil
			}
			return nil, status.Error(codes.FailedPrecondition, "group request already handled with a different result")
		}
		return nil, normalizeGroupWriteError(err, "failed to handle group request")
	}
	if !changed {
		return groupHandleResponse(&request, true), nil
	}

	notifyType := NotifyGroupReject
	if in.HandleResult == groupRequestAccepted {
		notifyType = NotifyGroupAccept
	}
	emitCommonNotify(l.ctx, l.ServiceContext, &mq.CommonNotify{
		NotifyType: notifyType, ReceiverId: request.ReqID, ActorId: in.ActorUid,
		BizId: fmt.Sprintf("group.handle:%d:%d", request.ID, in.HandleResult), GroupId: strconv.FormatUint(request.GroupID, 10), Content: in.HandleMsg,
	})
	request.HandleResult = intPtr(int(in.HandleResult))
	return groupHandleResponse(&request, false), nil
}

func groupHandleRequestID(in *social.GroupPutInHandleReq) (uint64, error) {
	legacy := uint64(0)
	if in.GroupReqId > 0 {
		legacy = uint64(in.GroupReqId)
	}
	if in.RequestId != 0 && legacy != 0 && in.RequestId != legacy {
		return 0, status.Error(codes.InvalidArgument, "request id does not match legacy group request id")
	}
	if in.RequestId != 0 {
		return in.RequestId, nil
	}
	return legacy, nil
}

func (l *GroupPutInHandleLogic) requireCurrentGroupApprover(groupID, actor uint64) error {
	if _, err := loadNormalGroup(l.DB.WithContext(l.ctx), groupID); err != nil {
		return normalizeGroupWriteError(err, "failed to validate group")
	}
	member, err := loadGroupMember(l.DB.WithContext(l.ctx), groupID, actor)
	if err != nil {
		return status.Error(codes.Internal, "failed to validate group approver")
	}
	if member == nil || (member.RoleLevel != 1 && member.RoleLevel != 2) {
		return status.Error(codes.PermissionDenied, "only a current group owner or administrator may handle requests")
	}
	return nil
}

func validateGroupHandleActor(actorUID, legacyUID string) (uint64, error) {
	if actorUID == "" {
		return 0, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if legacyUID != "" && legacyUID != actorUID {
		return 0, status.Error(codes.PermissionDenied, "actor uid does not match handle uid")
	}
	return parsePositiveID(actorUID, "actor uid")
}

func groupHandleResponse(request *objects.GroupRequest, idempotent bool) *social.GroupPutInHandleResp {
	result := int32(0)
	if request.HandleResult != nil {
		result = int32(*request.HandleResult)
	}
	return &social.GroupPutInHandleResp{
		GroupId: strconv.FormatUint(request.GroupID, 10), ReqId: request.ReqID,
		RequestId: request.ID, HandleResult: result, Idempotent: idempotent,
	}
}
