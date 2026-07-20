package logic

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MarkNotificationsReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记通知已读(单条/批量/全部)
func NewMarkNotificationsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkNotificationsReadLogic {
	return &MarkNotificationsReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkNotificationsReadLogic) MarkNotificationsRead(req *types.MarkNotificationsReadReq) (resp *types.MarkNotificationsReadResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	if parsed, parseErr := strconv.ParseUint(uid, 10, 64); parseErr != nil || parsed == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	ids, err := parseNotificationIDs(req.Ids)
	if err != nil {
		return nil, err
	}
	targets := make([]*im.NotificationReadTarget, 0, len(req.Targets))
	for _, target := range req.Targets {
		targets = append(targets, &im.NotificationReadTarget{NotifyType: target.NotifyType, BizId: target.BizId})
	}

	res, err := l.svcCtx.IM.MarkNotificationsRead(l.ctx, &im.MarkNotificationsReadReq{
		ReceiverId:  uid,
		Ids:         ids,
		NotifyTypes: req.NotifyTypes,
		BizIds:      req.BizIds,
		Targets:     targets,
	})
	if err != nil {
		return nil, err
	}

	return &types.MarkNotificationsReadResp{Affected: res.Affected, UnreadCount: res.UnreadCount}, nil
}

func parseNotificationIDs(values []string) ([]uint64, error) {
	ids := make([]uint64, 0, len(values))
	for _, value := range values {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 || strconv.FormatUint(parsed, 10) != value {
			return nil, status.Error(codes.InvalidArgument, "notification ids must be canonical positive decimal strings")
		}
		ids = append(ids, parsed)
	}
	return ids, nil
}
