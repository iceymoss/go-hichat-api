package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GroupSetAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupSetAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSetAdminLogic {
	return &GroupSetAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupSetAdminLogic) GroupSetAdmin(in *social.GroupSetAdminReq) (*social.GroupSetAdminResp, error) {
	actor, err := validateScopedActor(in.ActorUid, in.UserId)
	if err != nil {
		return nil, err
	}
	role := int(constants.AtLargeGroupRoleLevel)
	if in.IsAdmin {
		role = int(constants.ManagerGroupRoleLevel)
	}

	// 不允许操作群主本人
	filtered := make([]string, 0, len(in.MemberIds))
	for _, id := range in.MemberIds {
		if id == "" || id == actor {
			continue
		}
		filtered = append(filtered, id)
	}
	groupID, parseErr := parsePositiveID(in.GroupId, "group id")
	if parseErr != nil {
		return nil, parseErr
	}
	ownerID, parseErr := parsePositiveID(actor, "owner uid")
	if parseErr != nil {
		return nil, parseErr
	}
	memberIDs := make([]uint64, 0, len(filtered))
	seen := make(map[uint64]struct{}, len(filtered))
	for _, id := range filtered {
		memberID, parseErr := parsePositiveID(id, "member uid")
		if parseErr != nil {
			return nil, parseErr
		}
		if _, duplicate := seen[memberID]; duplicate {
			continue
		}
		seen[memberID] = struct{}{}
		memberIDs = append(memberIDs, memberID)
	}
	var changedIDs []uint64
	err = transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
		attemptChanged := make([]uint64, 0, len(memberIDs))
		attemptConverged := make([]uint64, 0, len(memberIDs))
		query := tx
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var currentGroup objects.Group
		if err := query.First(&currentGroup, groupID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return status.Error(codes.NotFound, "group not found")
			}
			return err
		}
		if currentGroup.Status != nil && *currentGroup.Status != 0 {
			return status.Error(codes.FailedPrecondition, "group is unavailable")
		}
		if currentGroup.CreatorUID != ownerID {
			return status.Error(codes.PermissionDenied, "only the current group owner may change administrators")
		}
		members, err := loadGroupMembersLocked(tx, groupID, memberIDs...)
		if err != nil {
			return err
		}
		for _, memberID := range memberIDs {
			member := members[memberID]
			if member == nil || member.RoleLevel == int(constants.CreatorGroupRoleLevel) {
				continue
			}
			attemptConverged = append(attemptConverged, memberID)
			if member.RoleLevel == role {
				continue
			}
			result := tx.Model(&objects.GroupMember{}).Where("id = ?", member.ID).Updates(map[string]any{
				"role_level": role, "operator_uid": ownerID,
			})
			if result.Error != nil {
				return result.Error
			}
			attemptChanged = append(attemptChanged, memberID)
		}
		if err := convergeAdminReceipts(tx, groupID, attemptConverged, in.IsAdmin); err != nil {
			return err
		}
		changedIDs = attemptChanged
		return nil
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to update member roles")
	}

	// 公共通知：被设为/取消管理员的成员收到通知
	notifyType := NotifyGroupAdminSet
	if !in.IsAdmin {
		notifyType = NotifyGroupAdminUnset
	}
	for _, memberID := range changedIDs {
		id := fmt.Sprintf("%d", memberID)
		emitCommonNotify(l.ctx, l.svcCtx, &mq.CommonNotify{
			NotifyType: notifyType,
			ReceiverId: id,
			ActorId:    actor,
			BizId:      fmt.Sprintf("%s:%s:%s:%d", notifyType, in.GroupId, id, time.Now().Unix()),
			GroupId:    in.GroupId,
		})
	}

	return &social.GroupSetAdminResp{}, nil
}
