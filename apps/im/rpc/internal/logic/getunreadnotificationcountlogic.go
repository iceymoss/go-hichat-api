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

type GetUnreadNotificationCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadNotificationCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadNotificationCountLogic {
	return &GetUnreadNotificationCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUnreadNotificationCount 公共通知通道：接收者未读数
func (l *GetUnreadNotificationCountLogic) GetUnreadNotificationCount(in *im.GetUnreadNotificationCountReq) (*im.GetUnreadNotificationCountResp, error) {
	if parsed, err := strconv.ParseUint(in.ReceiverId, 10, 64); err != nil || parsed == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid receiver identity")
	}
	cnt, err := l.svcCtx.NotificationModel.CountUnread(l.ctx, in.ReceiverId)
	if err != nil {
		return nil, err
	}
	return &im.GetUnreadNotificationCountResp{Count: cnt}, nil
}
