package logic

import (
	"context"
	"strconv"

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
	filtered := len(in.NotifyTypes) > 0 || len(in.BizIds) > 0
	if filtered && (len(in.NotifyTypes) == 0 || len(in.BizIds) == 0) {
		return nil, status.Error(codes.InvalidArgument, "notify types and biz ids must be provided together")
	}
	if filtered && len(in.Ids) > 0 {
		return nil, status.Error(codes.InvalidArgument, "ids and business filters cannot be combined")
	}
	if len(in.NotifyTypes) > 100 || len(in.BizIds) > 100 {
		return nil, status.Error(codes.InvalidArgument, "too many notification filters")
	}
	for _, value := range append(append([]string{}, in.NotifyTypes...), in.BizIds...) {
		if value == "" {
			return nil, status.Error(codes.InvalidArgument, "notification filters cannot be empty")
		}
	}
	var affected int64
	var err error
	if filtered {
		affected, err = l.svcCtx.NotificationModel.MarkReadByBusiness(l.ctx, in.ReceiverId, in.NotifyTypes, in.BizIds)
	} else {
		affected, err = l.svcCtx.NotificationModel.MarkRead(l.ctx, in.ReceiverId, in.Ids)
	}
	if err != nil {
		return nil, err
	}
	return &im.MarkNotificationsReadResp{Affected: affected}, nil
}
