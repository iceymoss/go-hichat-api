package relay

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type notificationModelStub struct {
	sent, failed bool
	attempts     int
	next         *time.Time
	dead         bool
}

func (m *notificationModelStub) InsertTx(*gorm.DB, *objects.SocialNotificationOutbox) error {
	return nil
}
func (m *notificationModelStub) ListDue(context.Context, time.Time, int) ([]*objects.SocialNotificationOutbox, error) {
	return nil, nil
}
func (m *notificationModelStub) MarkSent(context.Context, uint64, time.Time) error {
	m.sent = true
	return nil
}
func (m *notificationModelStub) MarkFailed(_ context.Context, _ uint64, a int, n *time.Time, _ string, d bool) error {
	m.failed = true
	m.attempts = a
	m.next = n
	m.dead = d
	return nil
}
func (m *notificationModelStub) CountByStatus(context.Context, int) (int64, error)   { return 0, nil }
func (m *notificationModelStub) ReplayDead(context.Context, []uint64) (int64, error) { return 0, nil }

type notifyClientStub struct {
	err     error
	message *mq.CommonNotify
}

func (c *notifyClientStub) Push(m *mq.CommonNotify) error { c.message = m; return c.err }

func TestNotificationRelayDeliver(t *testing.T) {
	row := func() *objects.SocialNotificationOutbox {
		return &objects.SocialNotificationOutbox{ID: 7, NotifyType: "friend.apply", ReceiverID: "2", ActorID: "1", BizID: "friend:9:apply", Payload: `{"requestType":"friend","requestId":9,"result":0,"content":"hello"}`, CreatedAt: time.Now()}
	}
	t.Run("success", func(t *testing.T) {
		m := &notificationModelStub{}
		c := &notifyClientStub{}
		NewNotification(&svc.ServiceContext{SocialNotificationOutboxModel: m, SocialRequestNotificationClient: c}).deliver(context.Background(), row())
		require.True(t, m.sent)
		require.Equal(t, uint64(7), c.message.EventId)
		require.Equal(t, "hello", c.message.Content)
	})
	t.Run("retry", func(t *testing.T) {
		m := &notificationModelStub{}
		c := &notifyClientStub{err: errors.New("kafka down")}
		NewNotification(&svc.ServiceContext{SocialNotificationOutboxModel: m, SocialRequestNotificationClient: c}).deliver(context.Background(), row())
		require.True(t, m.failed)
		require.Equal(t, 1, m.attempts)
		require.NotNil(t, m.next)
		require.False(t, m.dead)
	})
	t.Run("dead", func(t *testing.T) {
		m := &notificationModelStub{}
		c := &notifyClientStub{err: errors.New("kafka down")}
		r := row()
		r.Attempts = 9
		NewNotification(&svc.ServiceContext{SocialNotificationOutboxModel: m, SocialRequestNotificationClient: c}).deliver(context.Background(), r)
		require.True(t, m.dead)
		require.Nil(t, m.next)
	})
	t.Run("bad payload", func(t *testing.T) {
		m := &notificationModelStub{}
		c := &notifyClientStub{}
		r := row()
		r.Payload = "{"
		NewNotification(&svc.ServiceContext{SocialNotificationOutboxModel: m, SocialRequestNotificationClient: c}).deliver(context.Background(), r)
		require.True(t, m.dead)
		require.Nil(t, c.message)
	})
	t.Run("inconsistent payload", func(t *testing.T) {
		m := &notificationModelStub{}
		c := &notifyClientStub{}
		r := row()
		r.Payload = `{"requestType":"group","requestId":9,"result":1,"groupId":"1"}`
		NewNotification(&svc.ServiceContext{SocialNotificationOutboxModel: m, SocialRequestNotificationClient: c}).deliver(context.Background(), r)
		require.True(t, m.dead)
		require.Nil(t, c.message)
	})
}

func TestNotificationRelayAcceptsRequestInvalidationEvents(t *testing.T) {
	tests := []struct {
		name string
		row  *objects.SocialNotificationOutbox
	}{
		{name: "other administrator resolved", row: &objects.SocialNotificationOutbox{ID: 8, NotifyType: "group.request.resolved", ReceiverID: "2", ActorID: "1", BizID: "group:9:resolved", GroupID: "7", Payload: `{"requestType":"group","requestId":9,"result":1,"groupId":"7"}`, CreatedAt: time.Now()}},
		{name: "invitation invalidated", row: &objects.SocialNotificationOutbox{ID: 9, NotifyType: "group.invite.invalidated", ReceiverID: "3", ActorID: "1", BizID: "group_invite:10:invalidated", GroupID: "7", Payload: `{"requestType":"group_invite","requestId":10,"result":3,"groupId":"7"}`, CreatedAt: time.Now()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &notificationModelStub{}
			client := &notifyClientStub{}
			NewNotification(&svc.ServiceContext{SocialNotificationOutboxModel: model, SocialRequestNotificationClient: client}).deliver(context.Background(), tt.row)
			require.True(t, model.sent)
			require.False(t, model.dead)
			require.Equal(t, tt.row.NotifyType, client.message.NotifyType)
			require.Equal(t, tt.row.BizID, client.message.BizId)
		})
	}
}
