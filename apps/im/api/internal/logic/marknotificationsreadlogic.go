package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"

	"github.com/zeromicro/go-zero/core/logx"
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
	uid, _ := l.ctx.Value(Identify).(string)

	res, err := l.svcCtx.IM.MarkNotificationsRead(l.ctx, &im.MarkNotificationsReadReq{
		ReceiverId: uid,
		Ids:        req.Ids,
	})
	if err != nil {
		return nil, err
	}

	return &types.MarkNotificationsReadResp{Affected: res.Affected}, nil
}
