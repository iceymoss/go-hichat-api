package logic

import (
	"context"
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

type GroupInvitationCreateLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupInvitationCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationCreateLogic {
	return &GroupInvitationCreateLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupInvitationCreateLogic) GroupInvitationCreate(in *social.GroupInvitationCreateReq) (*social.GroupInvitationCreateResp, error) {
	actor, err := parsePositiveID(in.ActorUid, "actor uid")
	if in.ActorUid == "" {
		return nil, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if err != nil {
		return nil, err
	}
	groupID, err := parsePositiveID(in.GroupId, "group id")
	if err != nil {
		return nil, err
	}
	invitee, err := parsePositiveID(in.InviteeUid, "invitee uid")
	if err != nil {
		return nil, err
	}
	if actor == invitee {
		return nil, status.Error(codes.InvalidArgument, "cannot invite yourself")
	}
	if err := requireNormalUser(l.ctx, l.User, in.ActorUid); err != nil {
		return nil, err
	}
	if err := requireNormalUser(l.ctx, l.User, in.InviteeUid); err != nil {
		return nil, err
	}

	var invitation objects.GroupInvitation
	err = transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		if _, err := loadNormalGroup(tx, groupID); err != nil {
			return err
		}
		inviter, err := loadGroupMember(tx, groupID, actor)
		if err != nil {
			return err
		}
		if inviter == nil {
			return status.Error(codes.PermissionDenied, "inviter must be a current group member")
		}
		existing, err := loadGroupMember(tx, groupID, invitee)
		if err != nil {
			return err
		}
		if existing != nil {
			return status.Error(codes.AlreadyExists, "invitee is already a group member")
		}
		now := time.Now()
		invitation = objects.GroupInvitation{
			GroupID: groupID, InviterUID: actor, InviteeUID: invitee, InviterRoleSnapshot: inviter.RoleLevel,
			Message: in.Message, Status: groupInvitationPending, CreatedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour),
		}
		return tx.Create(&invitation).Error
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to create group invitation")
	}
	emitCommonNotify(l.ctx, l.ServiceContext, &mq.CommonNotify{
		NotifyType: NotifyGroupInvite, ReceiverId: in.InviteeUid, ActorId: in.ActorUid,
		BizId: fmt.Sprintf("group.invite:%d", invitation.ID), GroupId: in.GroupId, Content: in.Message,
	})
	return &social.GroupInvitationCreateResp{Invitation: invitationProto(&invitation)}, nil
}

func invitationProto(invitation *objects.GroupInvitation) *social.GroupInvitation {
	handledAt := int64(0)
	if invitation.HandledAt != nil {
		handledAt = invitation.HandledAt.Unix()
	}
	return &social.GroupInvitation{
		Id: invitation.ID, GroupId: strconv.FormatUint(invitation.GroupID, 10),
		InviterUid: strconv.FormatUint(invitation.InviterUID, 10), InviteeUid: strconv.FormatUint(invitation.InviteeUID, 10),
		InviterRoleSnapshot: int32(invitation.InviterRoleSnapshot), Message: invitation.Message, Status: int32(invitation.Status),
		RejectReason: invitation.RejectReason, CreatedAt: invitation.CreatedAt.Unix(), HandledAt: handledAt, ExpiresAt: invitation.ExpiresAt.Unix(),
	}
}
