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

type ListNotificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListNotificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNotificationsLogic {
	return &ListNotificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListNotifications 公共通知通道：拉取接收者通知列表（分页，按 id 倒序）。
func (l *ListNotificationsLogic) ListNotifications(in *im.ListNotificationsReq) (*im.ListNotificationsResp, error) {
	if !rpcauth.CanonicalUID(in.ReceiverId) {
		return nil, status.Error(codes.InvalidArgument, "receiver identity must be a canonical positive decimal string")
	}
	if err := requireNotificationUser(l.ctx, l.svcCtx.RPCAuth, in.ReceiverId); err != nil {
		return nil, err
	}
	if err := validateNotificationPagination(in.Offset, in.Limit); err != nil {
		return nil, err
	}
	rows, err := l.svcCtx.NotificationModel.ListByReceiver(l.ctx, in.ReceiverId, in.UnreadOnly, int(in.Offset), int(in.Limit))
	if err != nil {
		return nil, err
	}

	list := make([]*im.Notification, 0, len(rows))
	for _, r := range rows {
		var createTime int64
		if !r.CreatedAt.IsZero() {
			createTime = r.CreatedAt.Unix()
		}
		list = append(list, &im.Notification{
			Id:         r.Id,
			ReceiverId: r.ReceiverId,
			NotifyType: r.NotifyType,
			BizId:      r.BizId,
			ActorId:    r.ActorId,
			GroupId:    r.GroupId,
			Title:      r.Title,
			Content:    r.Content,
			Payload:    r.Payload,
			IsRead:     int32(r.IsRead),
			CreateTime: createTime,
		})
	}

	return &im.ListNotificationsResp{List: list}, nil
}

func validateNotificationPagination(offset, limit int32) error {
	if offset < 0 || offset > 100000 || limit < 0 || limit > 100 {
		return status.Error(codes.InvalidArgument, "notification pagination requires offset between 0 and 100000 and limit <= 100")
	}
	return nil
}
