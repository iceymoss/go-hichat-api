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

type GroupTransferOwnerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGroupTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupTransferOwnerLogic {
	return &GroupTransferOwnerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GroupTransferOwnerLogic) GroupTransferOwner(in *social.GroupTransferOwnerReq) (*social.GroupTransferOwnerResp, error) {
	if in.NewOwnerId == "" || in.NewOwnerId == in.UserId {
		return nil, status.Error(codes.InvalidArgument, "new owner must differ from current owner")
	}

	now := time.Now()
	oldRole := int(constants.AtLargeGroupRoleLevel)
	// 默认保持管理员（微信类似体验）
	if in.KeepOldOwnerAsAdmin {
		oldRole = int(constants.ManagerGroupRoleLevel)
	}
	groupID, parseErr := parsePositiveID(in.GroupId, "group id")
	if parseErr != nil {
		return nil, parseErr
	}
	oldOwnerID, parseErr := parsePositiveID(in.UserId, "owner uid")
	if parseErr != nil {
		return nil, parseErr
	}
	newOwnerID, parseErr := parsePositiveID(in.NewOwnerId, "new owner uid")
	if parseErr != nil {
		return nil, parseErr
	}
	err := transactionWithSQLiteRetry(l.ctx, l.svcCtx.DB, func(tx *gorm.DB) error {
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
		if currentGroup.CreatorUID != oldOwnerID {
			return status.Error(codes.PermissionDenied, "only the current group owner may transfer ownership")
		}
		members, err := loadGroupMembersLocked(tx, groupID, oldOwnerID, newOwnerID)
		if err != nil {
			return err
		}
		if members[oldOwnerID] == nil || members[newOwnerID] == nil {
			return status.Error(codes.FailedPrecondition, "new owner must be a current group member")
		}
		if err := tx.Model(&objects.Group{}).Where("id = ?", groupID).Updates(map[string]any{
			"creator_uid": newOwnerID, "updated_at": now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&objects.GroupMember{}).Where("id = ?", members[newOwnerID].ID).
			Updates(map[string]any{"role_level": int(constants.CreatorGroupRoleLevel)}).Error; err != nil {
			return err
		}
		return tx.Model(&objects.GroupMember{}).Where("id = ?", members[oldOwnerID].ID).
			Updates(map[string]any{"role_level": oldRole}).Error
	})
	if err != nil {
		return nil, normalizeGroupWriteError(err, "failed to transfer group ownership")
	}

	// 公共通知：新群主收到"已成为群主"通知
	emitCommonNotify(l.ctx, l.svcCtx, &mq.CommonNotify{
		NotifyType: NotifyGroupOwnerTransferred,
		ReceiverId: in.NewOwnerId,
		ActorId:    in.UserId,
		BizId:      fmt.Sprintf("group.owner:%s:%s:%d", in.GroupId, in.NewOwnerId, now.Unix()),
		GroupId:    in.GroupId,
	})

	return &social.GroupTransferOwnerResp{}, nil
}
