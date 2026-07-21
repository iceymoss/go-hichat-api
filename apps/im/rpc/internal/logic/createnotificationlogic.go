package logic

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err := requireNotificationTask(l.ctx, l.svcCtx.RPCAuth); err != nil {
		return nil, err
	}
	if err := validateCreateNotification(in, time.Now()); err != nil {
		return nil, err
	}
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
		Id:          data.Id,
		Inserted:    inserted,
		AlreadyRead: data.IsRead == 1,
	}, nil
}

func validateCreateNotification(in *im.CreateNotificationReq, now time.Time) error {
	if in == nil {
		return status.Error(codes.InvalidArgument, "notification is required")
	}
	if !rpcauth.CanonicalUID(in.ReceiverId) {
		return status.Error(codes.InvalidArgument, "receiver identity must be a canonical positive decimal string")
	}
	if strings.TrimSpace(in.NotifyType) == "" || strings.TrimSpace(in.BizId) == "" {
		return status.Error(codes.InvalidArgument, "notification type and biz id are required")
	}
	if !validNotificationText(in.NotifyType, 64) || !validNotificationText(in.BizId, 128) ||
		!validOptionalUID(in.ActorId) || !validOptionalUID(in.GroupId) ||
		!validNotificationText(in.Title, 255) || !validNotificationText(in.Content, 512) ||
		!utf8.ValidString(in.Payload) || len(in.Payload) > 65535 || !validCreateTime(in.CreateTime, now) {
		return status.Error(codes.InvalidArgument, "notification fields exceed allowed bounds or contain invalid values")
	}
	return nil
}

func validCreateTime(createTime int64, now time.Time) bool {
	if createTime == 0 {
		return true
	}
	createdAt := time.Unix(createTime, 0)
	return !createdAt.Before(now.Add(-30*24*time.Hour)) && !createdAt.After(now.Add(5*time.Minute))
}

func validOptionalUID(uid string) bool {
	return uid == "" || (len(uid) <= 64 && rpcauth.CanonicalUID(uid))
}

func validNotificationText(value string, maxRunes int) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) <= maxRunes
}
