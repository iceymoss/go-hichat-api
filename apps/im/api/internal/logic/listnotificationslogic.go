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

type ListNotificationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 拉取当前用户的通知列表(公共通知通道)
func NewListNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNotificationsLogic {
	return &ListNotificationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNotificationsLogic) ListNotifications(req *types.ListNotificationsReq) (resp *types.ListNotificationsResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	if parsed, parseErr := strconv.ParseUint(uid, 10, 64); parseErr != nil || parsed == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}

	res, err := l.svcCtx.IM.ListNotifications(l.ctx, &im.ListNotificationsReq{
		ReceiverId: uid,
		UnreadOnly: req.UnreadOnly,
		Offset:     req.Offset,
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.NotificationItem, 0, len(res.List))
	for _, v := range res.List {
		list = append(list, types.NotificationItem{
			Id:         strconv.FormatUint(v.Id, 10),
			NotifyType: v.NotifyType,
			BizId:      v.BizId,
			ActorId:    v.ActorId,
			GroupId:    v.GroupId,
			Title:      v.Title,
			Content:    v.Content,
			Payload:    v.Payload,
			IsRead:     v.IsRead,
			CreateTime: v.CreateTime,
		})
	}

	return &types.ListNotificationsResp{List: list}, nil
}
