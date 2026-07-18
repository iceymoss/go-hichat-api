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

type GroupInviteLogic struct {
	ctx context.Context
	*svc.ServiceContext
	logx.Logger
}

func NewGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLogic {
	return &GroupInviteLogic{ctx: ctx, ServiceContext: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GroupInviteLogic) GroupInvite(in *social.GroupInviteReq) (*social.GroupInviteResp, error) {
	actorText := in.ActorUid
	if actorText == "" {
		actorText = in.UserId
	}
	if in.ActorUid == "" {
		return nil, status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if in.UserId != "" && in.UserId != in.ActorUid {
		return nil, status.Error(codes.PermissionDenied, "actor uid does not match user id")
	}
	actor, err := parsePositiveID(actorText, "actor uid")
	if err != nil {
		return nil, err
	}
	if err := requireNormalUser(l.ctx, l.User, actorText); err != nil {
		return nil, err
	}
	groupID, err := parsePositiveID(in.GroupId, "group id")
	if err != nil {
		return nil, err
	}
	invitees := make([]uint64, 0, len(in.FriendIds))
	seen := make(map[uint64]struct{}, len(in.FriendIds))
	for _, id := range in.FriendIds {
		invitee, err := parsePositiveID(id, "invitee uid")
		if err != nil {
			return nil, err
		}
		if invitee == actor {
			return nil, status.Error(codes.InvalidArgument, "cannot invite yourself")
		}
		if _, duplicate := seen[invitee]; duplicate {
			return nil, status.Error(codes.InvalidArgument, "duplicate invitee uid")
		}
		seen[invitee] = struct{}{}
		if err := requireNormalUser(l.ctx, l.User, id); err != nil {
			return nil, err
		}
		invitees = append(invitees, invitee)
	}
	if len(invitees) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one invitee is required")
	}

	invitations := make([]objects.GroupInvitation, 0, len(invitees))
	err = transactionWithSQLiteRetry(l.ctx, l.DB, func(tx *gorm.DB) error {
		invitations = make([]objects.GroupInvitation, 0, len(invitees))
		if _, err := loadNormalGroup(tx, groupID); err != nil {
			return err
		}
		member, err := loadGroupMember(tx, groupID, actor)
		if err != nil {
			return err
		}
		if member == nil {
			return status.Error(codes.PermissionDenied, "inviter must be a current group member")
		}
		for _, invitee := range invitees {
			existing, err := loadGroupMember(tx, groupID, invitee)
			if err != nil {
				return err
			}
			if existing != nil {
				return status.Error(codes.AlreadyExists, "invitee is already a group member")
			}
			now := time.Now()
			invitations = append(invitations, objects.GroupInvitation{
				GroupID: groupID, InviterUID: actor, InviteeUID: invitee, InviterRoleSnapshot: member.RoleLevel,
				Message: "", Status: groupInvitationPending, CreatedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour),
			})
		}
		if err := tx.Create(&invitations).Error; err != nil {
			return err
		}
		for i := range invitations {
			if err := createReceipt(tx, receiptTypeGroupInvite, invitations[i].ID, strconv.FormatUint(invitations[i].InviteeUID, 10), receiptKindInvite, 1, receiptPending, invitations[i].CreatedAt, false, nil); err != nil {
				return err
			}
			if err := emitRequestNotificationInTx(tx, l.ServiceContext, NotifyGroupInvite, receiptTypeGroupInvite, invitations[i].ID, strconv.FormatUint(invitations[i].InviteeUID, 10), actorText, groupID, receiptPending); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to create group invitations")
	}
	ids := make([]uint64, len(invitations))
	for i := range invitations {
		ids[i] = invitations[i].ID
	}
	return &social.GroupInviteResp{InvitationIds: ids}, nil
}
