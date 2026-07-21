package logic

import (
	"context"
	"time"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNotificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNotificationLogic {
	return &CreateNotificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateNotification 公共通知通道：落库一条通知（幂等）。
// 命中 (receiver_id, notify_type, biz_id) 唯一键则跳过，inserted=false，供消费侧判断是否首次（避免重复推送）。
func (l *CreateNotificationLogic) CreateNotification(in *im.CreateNotificationReq) (*im.CreateNotificationResp, error) {
	data := &models.Notification{
		ReceiverId: in.ReceiverId,
		NotifyType: in.NotifyType,
		BizId:      in.BizId,
		ActorId:    in.ActorId,
		GroupId:    in.GroupId,
		Title:      in.Title,
		Content:    in.Content,
		Payload:    in.Payload,
		IsRead:     0,
	}
	if in.CreateTime > 0 {
		data.CreatedAt = time.Unix(in.CreateTime, 0)
	}

	inserted, err := l.svcCtx.NotificationModel.Insert(l.ctx, data)
	if err != nil {
		l.Errorf("CreateNotification insert failed: receiver=%s type=%s biz=%s err=%v",
			in.ReceiverId, in.NotifyType, in.BizId, err)
		return nil, err
	}

	return &im.CreateNotificationResp{
		Id:       data.Id,
		Inserted: inserted,
	}, nil
}
