package msg_transfer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	dlqVersion             = 1
	dlqMalformedJSON       = "malformed_json"
	dlqRPCInvalidArgument  = "rpc_invalid_argument"
	maxNotificationBackoff = 30 * time.Second
)

type notificationCreator interface {
	CreateNotification(context.Context, *im.CreateNotificationReq, ...grpc.CallOption) (*im.CreateNotificationResp, error)
}

type notificationSender interface {
	Send(any) error
}

type deadLetterPublisher interface {
	Publish(context.Context, []byte) error
}

type notificationDeadLetter struct {
	Version        int    `json:"version"`
	Classification string `json:"classification"`
	RawEvent       string `json:"rawEvent"`
}

// CommonNotifyTransfer persists notifications through the owning IM service before best-effort online delivery.
type CommonNotifyTransfer struct {
	im         notificationCreator
	ws         notificationSender
	dlq        deadLetterPublisher
	shutdown   context.Context
	begin      func() (func(), bool)
	retryDelay func(int) time.Duration
}

func NewCommonNotifyTransfer(svcCtx *svc.ServiceContext) kq.ConsumeHandler {
	return &CommonNotifyTransfer{
		im:         svcCtx.Im,
		ws:         svcCtx.WsClient,
		dlq:        svcCtx.NotificationDLQ,
		shutdown:   svcCtx.ShutdownContext(),
		begin:      svcCtx.BeginNotification,
		retryDelay: notificationRetryDelay,
	}
}

func (m *CommonNotifyTransfer) Consume(ctx context.Context, key, value string) error {
	if m.begin != nil {
		done, ok := m.begin()
		if !ok {
			return context.Canceled
		}
		defer done()
	}
	ctx, cancel := notificationConsumeContext(ctx, m.shutdown)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	var in mq.CommonNotify
	if err := json.Unmarshal([]byte(value), &in); err != nil {
		logx.WithContext(ctx).Errorf("CommonNotifyTransfer malformed event: err=%v", err)
		return m.publishDeadLetter(ctx, value, dlqMalformedJSON)
	}
	if in.ReceiverId == "" {
		return m.publishDeadLetter(ctx, value, dlqRPCInvalidArgument)
	}
	if in.ReceiverId == in.ActorId {
		return nil
	}
	createTime := in.CreateTime
	if createTime == 0 {
		createTime = time.Now().Unix()
	}

	req := &im.CreateNotificationReq{
		ReceiverId: in.ReceiverId,
		NotifyType: in.NotifyType,
		BizId:      in.BizId,
		ActorId:    in.ActorId,
		GroupId:    in.GroupId,
		Title:      in.Title,
		Content:    in.Content,
		Payload:    in.Payload,
		CreateTime: createTime,
	}
	resp, err := m.createNotification(ctx, req, value)
	if err != nil {
		return err
	}
	if !resp.Inserted || resp.AlreadyRead {
		return nil
	}

	if err := m.ws.Send(websocket.Message{
		FrameType: websocket.FrameNoAck,
		Method:    "push.notify",
		FormId:    constants.SYSTEM_ROOT_UID,
		Data: &ws.Notify{
			NotifyType: in.NotifyType,
			ReceiverId: in.ReceiverId,
			ActorId:    in.ActorId,
			BizId:      in.BizId,
			GroupId:    in.GroupId,
			Title:      in.Title,
			Content:    in.Content,
			Payload:    in.Payload,
			CreateTime: createTime,
		},
	}); err != nil {
		logx.WithContext(ctx).Errorf("CommonNotifyTransfer websocket delivery failed: event=%d type=%s receiver=%s err=%v", in.EventId, in.NotifyType, in.ReceiverId, err)
	}
	return nil
}

func (m *CommonNotifyTransfer) createNotification(ctx context.Context, req *im.CreateNotificationReq, raw string) (*im.CreateNotificationResp, error) {
	for attempt := 0; ; attempt++ {
		resp, err := m.im.CreateNotification(rpcauth.WithTask(ctx), req)
		if err == nil && resp != nil {
			return resp, nil
		}
		if err == nil {
			logx.WithContext(ctx).Error("CommonNotifyTransfer IM RPC returned an empty response, blocking partition")
			if err := waitNotificationRetry(ctx, m.retryDelay(attempt)); err != nil {
				return nil, err
			}
			continue
		}
		if status.Code(err) == codes.InvalidArgument {
			if err := m.publishDeadLetter(ctx, raw, dlqRPCInvalidArgument); err != nil {
				return nil, err
			}
			return &im.CreateNotificationResp{}, nil
		}
		if code := status.Code(err); code == codes.Unauthenticated || code == codes.PermissionDenied {
			logx.WithContext(ctx).Errorf("ALERT CommonNotifyTransfer IM RPC authentication/configuration failure: err=%v", err)
		} else {
			logx.WithContext(ctx).Errorf("CommonNotifyTransfer IM RPC failed, blocking partition: err=%v", err)
		}
		if err := waitNotificationRetry(ctx, m.retryDelay(attempt)); err != nil {
			return nil, err
		}
	}
}

func (m *CommonNotifyTransfer) publishDeadLetter(ctx context.Context, raw, classification string) error {
	body, err := json.Marshal(notificationDeadLetter{Version: dlqVersion, Classification: classification, RawEvent: raw})
	if err != nil {
		return err
	}
	for attempt := 0; ; attempt++ {
		if err := m.dlq.Publish(ctx, body); err == nil {
			return nil
		} else {
			logx.WithContext(ctx).Errorf("CommonNotifyTransfer DLQ publish failed, blocking partition: classification=%s err=%v", classification, err)
		}
		if err := waitNotificationRetry(ctx, m.retryDelay(attempt)); err != nil {
			return err
		}
	}
}

func notificationConsumeContext(handlerCtx, shutdownCtx context.Context) (context.Context, context.CancelFunc) {
	if shutdownCtx == nil {
		shutdownCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(handlerCtx)
	stop := context.AfterFunc(shutdownCtx, cancel)
	// AfterFunc schedules asynchronously when shutdown has already happened, so
	// synchronously close this window before processing a prefetched message.
	if shutdownCtx.Err() != nil {
		cancel()
	}
	return ctx, func() {
		stop()
		cancel()
	}
}

func notificationRetryDelay(attempt int) time.Duration {
	delay := time.Second
	for i := 0; i < attempt && delay < maxNotificationBackoff; i++ {
		delay *= 2
	}
	if delay > maxNotificationBackoff {
		return maxNotificationBackoff
	}
	return delay
}

func waitNotificationRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
