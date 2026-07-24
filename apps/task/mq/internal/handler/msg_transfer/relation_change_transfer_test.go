package msg_transfer

import (
	"context"
	"errors"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/task/mq/mq"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type ensureStub struct {
	err     error
	request *im.EnsureGroupConversationReq
}

type relationWSRecorder struct {
	messages []websocket.Message
}

func (r *relationWSRecorder) Close() error   { return nil }
func (r *relationWSRecorder) Read(any) error { return nil }
func (r *relationWSRecorder) Send(value any) error {
	r.messages = append(r.messages, value.(websocket.Message))
	return nil
}

func TestValidGroupRelationEvent(t *testing.T) {
	tests := []struct {
		name  string
		event mq.RelationChangeTransfer
		valid bool
	}{
		{name: "valid", event: mq.RelationChangeTransfer{UserId: "7", GroupId: "9", Version: 1, Timestamp: 1}, valid: true},
		{name: "missing user", event: mq.RelationChangeTransfer{GroupId: "9", Version: 1, Timestamp: 1}},
		{name: "invalid group", event: mq.RelationChangeTransfer{UserId: "7", GroupId: "x", Version: 1, Timestamp: 1}},
		{name: "missing version", event: mq.RelationChangeTransfer{UserId: "7", GroupId: "9", Timestamp: 1}},
		{name: "missing timestamp", event: mq.RelationChangeTransfer{UserId: "7", GroupId: "9", Version: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { require.Equal(t, test.valid, validGroupRelationEvent(&test.event)) })
	}
}

func (s *ensureStub) EnsureGroupConversation(_ context.Context, in *im.EnsureGroupConversationReq, _ ...grpc.CallOption) (*im.EnsureGroupConversationResp, error) {
	s.request = in
	return &im.EnsureGroupConversationResp{}, s.err
}

func TestEnsureGroupMemberConversation(t *testing.T) {
	event := &mq.RelationChangeTransfer{UserId: "7", GroupId: "9"}
	t.Run("success", func(t *testing.T) {
		client := &ensureStub{}
		require.NoError(t, ensureGroupMemberConversation(context.Background(), client, event))
		require.Equal(t, "7", client.request.UserId)
		require.Equal(t, "9", client.request.GroupId)
	})
	t.Run("retryable failure", func(t *testing.T) {
		expected := errors.New("im unavailable")
		client := &ensureStub{err: expected}
		require.ErrorIs(t, ensureGroupMemberConversation(context.Background(), client, event), expected)
	})
}

func TestFriendAddedPushesRelationChangeToBothUsers(t *testing.T) {
	recorder := &relationWSRecorder{}
	consumer := &RelationChangeTransfer{BaseChatTransfer: NewBaseMsgChatTransfer(&svc.ServiceContext{WsClient: recorder})}
	consumer.pushRelationChanged(&mq.RelationChangeTransfer{
		EventType: constants.RelationEventFriendAdded,
		FriendA:   "1",
		FriendB:   "2",
	})

	require.Len(t, recorder.messages, 2)
	first, ok := recorder.messages[0].Data.(*ws.RelationChanged)
	require.True(t, ok)
	second, ok := recorder.messages[1].Data.(*ws.RelationChanged)
	require.True(t, ok)
	require.Equal(t, constants.RelationEventFriendAdded, first.EventType)
	require.Equal(t, []string{"1", "2"}, []string{first.ReceiverId, second.ReceiverId})
	require.Equal(t, []string{"2", "1"}, []string{first.PeerId, second.PeerId})
}
