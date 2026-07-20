package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MarkNotificationsReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkNotificationsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkNotificationsReadLogic {
	return &MarkNotificationsReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkNotificationsRead 公共通知通道：标记通知已读（ids 为空表示该接收者全部未读）。
func (l *MarkNotificationsReadLogic) MarkNotificationsRead(in *im.MarkNotificationsReadReq) (*im.MarkNotificationsReadResp, error) {
	if parsed, err := strconv.ParseUint(in.ReceiverId, 10, 64); err != nil || parsed == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid receiver identity")
	}
	legacyFiltered := len(in.NotifyTypes) > 0 || len(in.BizIds) > 0
	if legacyFiltered && (len(in.NotifyTypes) == 0 || len(in.BizIds) == 0 || len(in.NotifyTypes) != len(in.BizIds)) {
		return nil, status.Error(codes.InvalidArgument, "notify types and biz ids must be provided together")
	}
	if (legacyFiltered || len(in.Targets) > 0) && len(in.Ids) > 0 {
		return nil, status.Error(codes.InvalidArgument, "ids and business filters cannot be combined")
	}
	if len(in.Ids) > 0 {
		seenIDs := make(map[uint64]struct{}, len(in.Ids))
		dedupedIDs := in.Ids[:0]
		for _, id := range in.Ids {
			if id == 0 {
				return nil, status.Error(codes.InvalidArgument, "notification ids must be positive")
			}
			if _, ok := seenIDs[id]; ok {
				continue
			}
			seenIDs[id] = struct{}{}
			dedupedIDs = append(dedupedIDs, id)
		}
		in.Ids = dedupedIDs
		if len(in.Ids) > 100 {
			return nil, status.Error(codes.InvalidArgument, "too many notification ids")
		}
	}
	if legacyFiltered && len(in.Targets) > 0 {
		return nil, status.Error(codes.InvalidArgument, "legacy business filters and targets cannot be combined")
	}

	targets := make([]models.NotificationReadTarget, 0, len(in.Targets)+len(in.NotifyTypes))
	for _, target := range in.Targets {
		if target == nil {
			return nil, status.Error(codes.InvalidArgument, "notification target cannot be null")
		}
		modelTarget := models.NotificationReadTarget{NotifyType: target.NotifyType, BizId: target.BizId}
		if !canonicalSocialNotificationTarget(modelTarget.NotifyType, modelTarget.BizId) {
			return nil, status.Error(codes.InvalidArgument, "invalid social notification target")
		}
		targets = append(targets, modelTarget)
	}
	legacyTargets, err := legacyNotificationReadTargets(in.NotifyTypes, in.BizIds)
	if err != nil {
		return nil, err
	}
	targets = append(targets, legacyTargets...)
	seen := make(map[string]struct{}, len(targets))
	deduped := targets[:0]
	for _, target := range targets {
		key := target.NotifyType + "\x00" + target.BizId
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, target)
		if len(deduped) > 100 {
			return nil, status.Error(codes.InvalidArgument, "too many notification targets")
		}
	}

	var affected, unreadCount int64
	if len(deduped) > 0 {
		affected, unreadCount, err = l.svcCtx.NotificationModel.MarkReadByBusiness(l.ctx, in.ReceiverId, deduped)
	} else {
		affected, unreadCount, err = l.svcCtx.NotificationModel.MarkRead(l.ctx, in.ReceiverId, in.Ids)
	}
	if err != nil {
		return nil, err
	}
	return &im.MarkNotificationsReadResp{Affected: affected, UnreadCount: unreadCount}, nil
}

func legacyNotificationReadTargets(notifyTypes, bizIDs []string) ([]models.NotificationReadTarget, error) {
	if len(notifyTypes) != len(bizIDs) {
		return nil, status.Error(codes.InvalidArgument, "notify types and biz ids must be provided together")
	}
	targets := make([]models.NotificationReadTarget, 0, len(notifyTypes))
	for i := range notifyTypes {
		if notifyTypes[i] == "" || bizIDs[i] == "" {
			return nil, status.Error(codes.InvalidArgument, "legacy notification filters cannot be empty")
		}
		targets = append(targets, models.NotificationReadTarget{NotifyType: notifyTypes[i], BizId: bizIDs[i]})
	}
	return targets, nil
}

func canonicalSocialNotificationTarget(notifyType, bizID string) bool {
	parts := strings.Split(bizID, ":")
	if len(parts) != 3 || parts[1] == "" {
		return false
	}
	requestID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || requestID == 0 || strconv.FormatUint(requestID, 10) != parts[1] {
		return false
	}
	expected := map[string]string{
		"friend.apply":      "friend:apply",
		"friend.accept":     "friend:accept",
		"friend.reject":     "friend:reject",
		"group.apply":       "group:apply",
		"group.accept":      "group:accept",
		"group.reject":      "group:reject",
		"group.invalidated": "group:invalidated",
		"group.invite":      "group_invite:invite",
	}
	return expected[notifyType] == fmt.Sprintf("%s:%s", parts[0], parts[2])
}
