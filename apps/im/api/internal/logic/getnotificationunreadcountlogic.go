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

type GetNotificationUnreadCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 当前用户未读通知数
func NewGetNotificationUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNotificationUnreadCountLogic {
	return &GetNotificationUnreadCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNotificationUnreadCountLogic) GetNotificationUnreadCount() (resp *types.NotificationUnreadCountResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	if parsed, parseErr := strconv.ParseUint(uid, 10, 64); parseErr != nil || parsed == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}

	res, err := l.svcCtx.IM.GetUnreadNotificationCount(l.ctx, &im.GetUnreadNotificationCountReq{
		ReceiverId: uid,
	})
	if err != nil {
		return nil, err
	}

	return &types.NotificationUnreadCountResp{Count: res.Count}, nil
}
