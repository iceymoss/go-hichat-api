package relay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/zeromicro/go-zero/core/logx"
)

const notificationMaxAttempts = 10

type NotificationRelay struct{ svcCtx *svc.ServiceContext }

func NewNotification(s *svc.ServiceContext) *NotificationRelay { return &NotificationRelay{svcCtx: s} }
func (r *NotificationRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}
func (r *NotificationRelay) tick(ctx context.Context) {
	lock := db.GetRedisConn()
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return
	}
	token := hex.EncodeToString(tokenBytes)
	ok, err := lock.SetNX(ctx, "social:notification:relay:lock", token, 30*time.Second).Result()
	if err != nil || !ok {
		return
	}
	renewCtx, cancelRenew := context.WithCancel(ctx)
	defer cancelRenew()
	lost := make(chan struct{})
	go renewNotificationLock(renewCtx, lock, token, lost)
	defer lock.Eval(context.Background(), `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`, []string{"social:notification:relay:lock"}, token)
	rows, err := r.svcCtx.SocialNotificationOutboxModel.ListDue(ctx, time.Now(), 100)
	if err != nil {
		notificationFailures.WithLabelValues("list").Inc()
		logx.WithContext(ctx).Errorf("notification relay list due: %v", err)
		return
	}
	if pending, err := r.svcCtx.SocialNotificationOutboxModel.CountByStatus(ctx, 0); err == nil {
		notificationPending.Set(float64(pending))
	}
	if dead, err := r.svcCtx.SocialNotificationOutboxModel.CountByStatus(ctx, 2); err == nil {
		notificationDead.Set(float64(dead))
	}
	for _, row := range rows {
		select {
		case <-lost:
			return
		default:
		}
		r.deliver(ctx, row)
	}
}

func renewNotificationLock(ctx context.Context, lock *redis.Client, token string, lost chan<- struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := lock.Eval(ctx, `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`, []string{"social:notification:relay:lock"}, token, int64((30*time.Second)/time.Millisecond)).Int()
			if err != nil || result == 0 {
				close(lost)
				return
			}
		}
	}
}

func (r *NotificationRelay) deliver(ctx context.Context, row *objects.SocialNotificationOutbox) {
	var p struct {
		RequestType string `json:"requestType"`
		RequestID   uint64 `json:"requestId"`
		Result      int32  `json:"result"`
		Content     string `json:"content"`
		GroupID     string `json:"groupId"`
	}
	if err := json.Unmarshal([]byte(row.Payload), &p); err != nil {
		notificationFailures.WithLabelValues("payload").Inc()
		if markErr := r.svcCtx.SocialNotificationOutboxModel.MarkFailed(ctx, row.ID, row.Attempts+1, nil, "invalid payload", true); markErr != nil {
			logx.WithContext(ctx).Errorf("notification relay mark dead: event=%d err=%v", row.ID, markErr)
		}
		return
	}
	if p.RequestType == "" || p.RequestID == 0 {
		notificationFailures.WithLabelValues("payload").Inc()
		if markErr := r.svcCtx.SocialNotificationOutboxModel.MarkFailed(ctx, row.ID, row.Attempts+1, nil, "invalid payload fields", true); markErr != nil {
			logx.WithContext(ctx).Errorf("notification relay mark dead: event=%d err=%v", row.ID, markErr)
		}
		return
	}
	if !validNotification(row, p.RequestType, p.RequestID, p.Result, p.GroupID) {
		notificationFailures.WithLabelValues("payload").Inc()
		if markErr := r.svcCtx.SocialNotificationOutboxModel.MarkFailed(ctx, row.ID, row.Attempts+1, nil, "inconsistent notification fields", true); markErr != nil {
			logx.WithContext(ctx).Errorf("notification relay mark dead: event=%d err=%v", row.ID, markErr)
		}
		return
	}
	ev := &mq.CommonNotify{EventId: row.ID, NotifyType: row.NotifyType, ReceiverId: row.ReceiverID, ActorId: row.ActorID, BizId: row.BizID, GroupId: row.GroupID, Payload: row.Payload, RequestType: p.RequestType, RequestId: p.RequestID, Result: p.Result, Content: p.Content, CreateTime: row.CreatedAt.Unix()}
	if err := r.svcCtx.SocialRequestNotificationClient.Push(ev); err != nil {
		notificationFailures.WithLabelValues("publish").Inc()
		attempts := row.Attempts + 1
		dead := attempts >= notificationMaxAttempts
		var next *time.Time
		if !dead {
			n := time.Now().Add(time.Duration(math.Min(math.Pow(2, float64(attempts-1)), 300)) * time.Second)
			next = &n
		}
		if markErr := r.svcCtx.SocialNotificationOutboxModel.MarkFailed(ctx, row.ID, attempts, next, err.Error(), dead); markErr != nil {
			logx.WithContext(ctx).Errorf("notification relay mark failed: event=%d err=%v", row.ID, markErr)
		}
		return
	}
	if err := r.svcCtx.SocialNotificationOutboxModel.MarkSent(ctx, row.ID, time.Now()); err != nil {
		logx.WithContext(ctx).Errorf("notification relay mark sent: event=%d err=%v", row.ID, err)
	}
	notificationLatency.Observe(time.Since(row.CreatedAt).Seconds())
}

func validNotification(row *objects.SocialNotificationOutbox, requestType string, requestID uint64, result int32, groupID string) bool {
	if requestType == "friend" {
		if row.GroupID != "" || groupID != "" || (result != 0 && result != 1 && result != 2) {
			return false
		}
	} else {
		if result < 0 || result > 3 || (requestType == "group_invite" && result != 0) {
			return false
		}
		parsedGroupID, err := strconv.ParseUint(groupID, 10, 64)
		if err != nil || parsedGroupID == 0 {
			return false
		}
	}
	phase := "apply"
	if requestType == "group_invite" {
		phase = "invite"
	} else if result == 1 {
		phase = "accept"
	} else if result == 2 {
		phase = "reject"
	} else if result == 3 {
		phase = "invalidated"
	}
	expected := map[string]map[string]string{
		"friend":       {"apply": "friend.apply", "accept": "friend.accept", "reject": "friend.reject"},
		"group":        {"apply": "group.apply", "accept": "group.accept", "reject": "group.reject", "invalidated": "group.invalidated"},
		"group_invite": {"invite": "group.invite"},
	}
	types, ok := expected[requestType]
	if !ok || types[phase] != row.NotifyType || row.BizID != fmt.Sprintf("%s:%d:%s", requestType, requestID, phase) {
		return false
	}
	if requestType != "friend" && (row.GroupID == "" || groupID != row.GroupID) {
		return false
	}
	return row.ReceiverID != ""
}
