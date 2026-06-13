package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"

	"github.com/zeromicro/go-zero/core/logx"
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
	uid, _ := l.ctx.Value(Identify).(string)

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
			Id:         v.Id,
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
