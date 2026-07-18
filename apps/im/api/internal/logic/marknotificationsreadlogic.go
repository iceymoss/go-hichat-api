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

	res, err := l.svcCtx.IM.MarkNotificationsRead(l.ctx, &im.MarkNotificationsReadReq{
		ReceiverId:  uid,
		Ids:         req.Ids,
		NotifyTypes: req.NotifyTypes,
		BizIds:      req.BizIds,
	})
	if err != nil {
		return nil, err
	}

	return &types.MarkNotificationsReadResp{Affected: res.Affected}, nil
}
