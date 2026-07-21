package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
	affected, err := l.svcCtx.NotificationModel.MarkRead(l.ctx, in.ReceiverId, in.Ids)
	if err != nil {
		return nil, err
	}
	return &im.MarkNotificationsReadResp{Affected: affected}, nil
}
