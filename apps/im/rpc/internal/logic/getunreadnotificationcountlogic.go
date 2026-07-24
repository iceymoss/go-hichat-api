package logic

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

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
	if !rpcauth.CanonicalUID(in.ReceiverId) {
		return nil, status.Error(codes.InvalidArgument, "receiver identity must be a canonical positive decimal string")
	}
	if err := requireNotificationUser(l.ctx, l.svcCtx.RPCAuth, in.ReceiverId); err != nil {
		return nil, err
	}
	cnt, err := l.svcCtx.NotificationModel.CountUnread(l.ctx, in.ReceiverId)
	if err != nil {
		return nil, err
	}
	return &im.GetUnreadNotificationCountResp{Count: cnt}, nil
}
