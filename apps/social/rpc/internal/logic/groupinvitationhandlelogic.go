package logic

import (
	"context"
	"errors"
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

var errInvitationCASMiss = errors.New("group invitation CAS missed")

type GroupInvitationHandleLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupInvitationHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationHandleLogic {
	return &GroupInvitationHandleLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupInvitationHandleLogic) GroupInvitationHandle(in *social.GroupInvitationHandleReq) (*social.GroupInvitationHandleResp, error) {
	if in.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "invitation id must be positive")
	}
	if in.Result != 1 && in.Result != 2 {
		return nil, status.Error(codes.InvalidArgument, "result must be 1 or 2")
	}
	if in.ActorUid == "" {
		return nil, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	actor, err := parsePositiveID(in.ActorUid, "actor uid")
	if err != nil {
		return nil, err
	}
	if err := requireNormalUser(l.ctx, l.User, in.ActorUid); err != nil {
		return nil, err
	}
	var initial objects.GroupInvitation
	if err := l.DB.WithContext(l.ctx).First(&initial, in.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "group invitation not found")
		}
		return nil, status.Error(codes.Internal, "failed to load group invitation")
	}
	if initial.InviteeUID != actor {
		return nil, status.Error(codes.PermissionDenied, "only the invitee may handle this invitation")
	}

	var response *social.GroupInvitationHandleResp
	createdApproval := false
	err = transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		response = nil
		createdApproval = false
		var group objects.Group
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&group, initial.GroupID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "group not found")
			}
			return err
		}
		var invitation objects.GroupInvitation
		if err := tx.First(&invitation, in.Id).Error; err != nil {
			return err
		}
		if invitation.InviteeUID != actor {
			return status.Error(codes.PermissionDenied, "only the invitee may handle this invitation")
		}
		if invitation.Status != groupInvitationPending {
			return l.invitationTerminalResponse(tx, &invitation, in.Result, &response)
		}
		now := time.Now()
		if !invitation.ExpiresAt.After(now) {
			result := tx.Model(&objects.GroupInvitation{}).
				Where("id = ? AND invitee_uid = ? AND status = ?", invitation.ID, actor, groupInvitationPending).
				Updates(map[string]any{"status": groupInvitationExpired, "handled_at": now})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errInvitationCASMiss
			}
			response = invitationHandleResponse(&invitation, groupInvitationExpired, "expired", 0, false)
			return nil
		}
		if group.Status != nil && *group.Status != 0 {
			return invalidateInvitation(tx, &invitation, actor, now, "group is unavailable", &response)
		}
		if in.Result == 2 {
			result := tx.Model(&objects.GroupInvitation{}).
				Where("id = ? AND invitee_uid = ? AND status = ? AND expires_at > ?", invitation.ID, actor, groupInvitationPending, now).
				Updates(map[string]any{"status": groupInvitationRejected, "reject_reason": in.HandleMsg, "handled_at": now})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errInvitationCASMiss
			}
			response = invitationHandleResponse(&invitation, groupInvitationRejected, "rejected", 0, false)
			return nil
		}

		inviter, err := loadGroupMemberLocked(tx, invitation.GroupID, invitation.InviterUID)
		if err != nil {
			return err
		}
		if inviter == nil {
			return invalidateInvitation(tx, &invitation, actor, now, "inviter is no longer a group member", &response)
		}
		currentMember, err := loadGroupMemberLocked(tx, invitation.GroupID, actor)
		if err != nil {
			return err
		}
		if err := acceptCurrentInvitation(tx, &invitation, actor, now); err != nil {
			return err
		}
		if err := invalidateOtherInvitations(tx, &invitation, now); err != nil {
			return err
		}
		if currentMember != nil {
			if err := resolvePendingGroupRequests(tx, invitation.GroupID, actor, now, 2); err != nil {
				return err
			}
			response = invitationHandleResponse(&invitation, groupInvitationAccepted, "joined", 0, false)
			return nil
		}
		if inviter.RoleLevel == 1 || inviter.RoleLevel == 2 {
			if _, err := createGroupMemberAndOutbox(tx, l.ServiceContext, invitation.GroupID, actor, 0, 2, &invitation.InviterUID, invitation.InviterUID); err != nil {
				return err
			}
			if err := resolvePendingGroupRequests(tx, invitation.GroupID, actor, now, 2); err != nil {
				return err
			}
			response = invitationHandleResponse(&invitation, groupInvitationAccepted, "joined", 0, false)
			return nil
		}

		pending := groupRequestPending
		joinSource := 2
		request := objects.GroupRequest{
			ReqID: strconv.FormatUint(actor, 10), GroupID: invitation.GroupID, ReqMsg: invitation.Message, ReqTime: &now,
			JoinSource: &joinSource, InviterUserID: &invitation.InviterUID, HandleResult: &pending,
			SourceType: 2, SourceInvitationID: &invitation.ID,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&request)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			if err := tx.Where("source_invitation_id = ?", invitation.ID).First(&request).Error; err != nil {
				return err
			}
		}
		createdApproval = result.RowsAffected == 1
		response = invitationHandleResponse(&invitation, groupInvitationAccepted, "pending_approval", request.ID, false)
		return nil
	})
	if err != nil {
		if errors.Is(err, errInvitationCASMiss) {
			var latest objects.GroupInvitation
			if loadErr := l.DB.WithContext(l.ctx).First(&latest, in.Id).Error; loadErr != nil {
				return nil, status.Error(codes.Internal, "failed to reload group invitation")
			}
			if terminalErr := l.invitationTerminalResponse(l.DB.WithContext(l.ctx), &latest, in.Result, &response); terminalErr != nil {
				return nil, terminalErr
			}
			return response, nil
		}
		return nil, normalizeGroupWriteError(err, "failed to handle group invitation")
	}
	if createdApproval {
		notifyGroupAdminsFromDB(l.ctx, l.ServiceContext, initial.GroupID, actor, response.GroupRequestId, initial.Message)
	}
	return response, nil
}

func acceptCurrentInvitation(tx *gorm.DB, invitation *objects.GroupInvitation, actor uint64, now time.Time) error {
	result := tx.Model(&objects.GroupInvitation{}).
		Where("id = ? AND invitee_uid = ? AND status = ? AND expires_at > ?", invitation.ID, actor, groupInvitationPending, now).
		Updates(map[string]any{"status": groupInvitationAccepted, "handled_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errInvitationCASMiss
	}
	return nil
}

func invalidateOtherInvitations(tx *gorm.DB, invitation *objects.GroupInvitation, now time.Time) error {
	return tx.Model(&objects.GroupInvitation{}).
		Where("id <> ? AND group_id = ? AND invitee_uid = ? AND status = ?", invitation.ID, invitation.GroupID, invitation.InviteeUID, groupInvitationPending).
		Updates(map[string]any{"status": groupInvitationInvalidated, "handled_at": now}).Error
}

func invalidateInvitation(tx *gorm.DB, invitation *objects.GroupInvitation, actor uint64, now time.Time, _ string, response **social.GroupInvitationHandleResp) error {
	result := tx.Model(&objects.GroupInvitation{}).
		Where("id = ? AND invitee_uid = ? AND status = ?", invitation.ID, actor, groupInvitationPending).
		Updates(map[string]any{"status": groupInvitationInvalidated, "handled_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errInvitationCASMiss
	}
	*response = invitationHandleResponse(invitation, groupInvitationInvalidated, "invalidated", 0, false)
	return nil
}

func (l *GroupInvitationHandleLogic) invitationTerminalResponse(db *gorm.DB, invitation *objects.GroupInvitation, requestedResult int32, response **social.GroupInvitationHandleResp) error {
	if (requestedResult == 1 && invitation.Status == groupInvitationAccepted) || (requestedResult == 2 && invitation.Status == groupInvitationRejected) {
		joinState := "rejected"
		var requestID uint64
		if invitation.Status == groupInvitationAccepted {
			var request objects.GroupRequest
			requestErr := db.Where("source_invitation_id = ?", invitation.ID).First(&request).Error
			if requestErr == nil {
				requestID = request.ID
			} else if requestErr != gorm.ErrRecordNotFound {
				return requestErr
			}
			member, err := loadGroupMember(db, invitation.GroupID, invitation.InviteeUID)
			if err != nil {
				return err
			}
			if member != nil {
				*response = invitationHandleResponse(invitation, invitation.Status, "joined", requestID, true)
				return nil
			}
			if requestErr == nil {
				if request.HandleResult == nil || *request.HandleResult == groupRequestPending {
					joinState = "pending_approval"
				} else if *request.HandleResult == groupRequestAccepted {
					return status.Error(codes.FailedPrecondition, "accepted group approval is missing membership")
				} else if *request.HandleResult == groupRequestRejected {
					joinState = "approval_rejected"
				} else {
					return status.Error(codes.Internal, "group approval has invalid result")
				}
			} else {
				joinState = "joined"
			}
		}
		*response = invitationHandleResponse(invitation, invitation.Status, joinState, requestID, true)
		return nil
	}
	if invitation.Status == groupInvitationExpired {
		*response = invitationHandleResponse(invitation, invitation.Status, "expired", 0, true)
		return nil
	}
	if invitation.Status == groupInvitationInvalidated {
		*response = invitationHandleResponse(invitation, invitation.Status, "invalidated", 0, true)
		return nil
	}
	return status.Error(codes.FailedPrecondition, "group invitation already handled with a different result")
}

func invitationHandleResponse(invitation *objects.GroupInvitation, invitationStatus int, joinState string, requestID uint64, idempotent bool) *social.GroupInvitationHandleResp {
	return &social.GroupInvitationHandleResp{
		InvitationId: invitation.ID, Status: int32(invitationStatus), JoinState: joinState,
		GroupRequestId: requestID, Idempotent: idempotent, GroupId: strconv.FormatUint(invitation.GroupID, 10),
	}
}
