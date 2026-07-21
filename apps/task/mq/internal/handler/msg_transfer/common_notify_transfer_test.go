package msg_transfer

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/websocket"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/ws"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type fakeNotificationCreator struct {
	requests []*im.CreateNotificationReq
	results  []notificationRPCResult
}

type notificationRPCResult struct {
	resp *im.CreateNotificationResp
	err  error
}

type failingNotificationCreator struct {
	called chan struct{}
}

func (f *failingNotificationCreator) CreateNotification(context.Context, *im.CreateNotificationReq, ...grpc.CallOption) (*im.CreateNotificationResp, error) {
	select {
	case f.called <- struct{}{}:
	default:
	}
	return nil, status.Error(codes.Unavailable, "down")
}

type blockingNotificationCreator struct {
	called chan struct{}
}

func (f *blockingNotificationCreator) CreateNotification(ctx context.Context, _ *im.CreateNotificationReq, _ ...grpc.CallOption) (*im.CreateNotificationResp, error) {
	close(f.called)
	<-ctx.Done()
	return nil, ctx.Err()
}

func (f *fakeNotificationCreator) CreateNotification(_ context.Context, req *im.CreateNotificationReq, _ ...grpc.CallOption) (*im.CreateNotificationResp, error) {
	f.requests = append(f.requests, proto.Clone(req).(*im.CreateNotificationReq))
	result := f.results[0]
	f.results = f.results[1:]
	return result.resp, result.err
}

type fakeNotificationSender struct {
	messages []any
	err      error
}

func (f *fakeNotificationSender) Send(message any) error {
	f.messages = append(f.messages, message)
	return f.err
}

type fakeDeadLetterPublisher struct {
	values []string
	errors []error
}

type blockingDeadLetterPublisher struct {
	called chan struct{}
}

func (f *blockingDeadLetterPublisher) Publish(ctx context.Context, _ []byte) error {
	close(f.called)
	<-ctx.Done()
	return ctx.Err()
}

func (f *fakeDeadLetterPublisher) Publish(_ context.Context, value []byte) error {
	f.values = append(f.values, string(value))
	if len(f.errors) == 0 {
		return nil
	}
	err := f.errors[0]
	f.errors = f.errors[1:]
	return err
}

func newTestNotificationTransfer(rpc *fakeNotificationCreator, sender *fakeNotificationSender, dlq *fakeDeadLetterPublisher) *CommonNotifyTransfer {
	return &CommonNotifyTransfer{im: rpc, ws: sender, dlq: dlq, shutdown: context.Background(), retryDelay: func(int) time.Duration { return 0 }}
}

func TestCommonNotifyTransferMapsRPCAndPushes(t *testing.T) {
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{{resp: &im.CreateNotificationResp{Id: math.MaxUint64, Inserted: true}}}}
	sender := &fakeNotificationSender{}
	transfer := newTestNotificationTransfer(rpc, sender, &fakeDeadLetterPublisher{})
	raw := `{"eventId":18446744073709551615,"receiverId":"18446744073709551615","notifyType":"group.apply","bizId":"18446744073709551615","actorId":"9223372036854775808","groupId":"18446744073709551614","title":"title","content":"content","payload":"{\"requestId\":18446744073709551615}","createTime":1710000000}`

	if err := transfer.Consume(context.Background(), "", raw); err != nil {
		t.Fatal(err)
	}
	if len(rpc.requests) != 1 {
		t.Fatalf("CreateNotification calls = %d, want 1", len(rpc.requests))
	}
	want := &im.CreateNotificationReq{ReceiverId: "18446744073709551615", NotifyType: "group.apply", BizId: "18446744073709551615", ActorId: "9223372036854775808", GroupId: "18446744073709551614", Title: "title", Content: "content", Payload: `{"requestId":18446744073709551615}`, CreateTime: 1710000000}
	if got := rpc.requests[0]; !proto.Equal(got, want) {
		t.Fatalf("request = %+v, want %+v", got, want)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("websocket sends = %d, want 1", len(sender.messages))
	}
	message := sender.messages[0].(websocket.Message)
	notify := message.Data.(*ws.Notify)
	if message.Method != "push.notify" || notify.ReceiverId != want.ReceiverId || notify.BizId != want.BizId || notify.CreateTime != want.CreateTime {
		t.Fatalf("websocket message = %+v notify=%+v", message, notify)
	}
}

func TestCommonNotifyTransferRetriesRPCThenSucceeds(t *testing.T) {
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{
		{err: status.Error(codes.Unavailable, "temporary")},
		{err: status.Error(codes.DeadlineExceeded, "temporary")},
		{resp: &im.CreateNotificationResp{Inserted: true}},
	}}
	sender := &fakeNotificationSender{}
	transfer := newTestNotificationTransfer(rpc, sender, &fakeDeadLetterPublisher{})

	if err := transfer.Consume(context.Background(), "", validNotificationJSON()); err != nil {
		t.Fatal(err)
	}
	if len(rpc.requests) != 3 || len(sender.messages) != 1 {
		t.Fatalf("RPC calls=%d websocket sends=%d, want 3 and 1", len(rpc.requests), len(sender.messages))
	}
}

func TestCommonNotifyTransferCancellationStopsRetry(t *testing.T) {
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{{err: status.Error(codes.Unavailable, "down")}}}
	transfer := &CommonNotifyTransfer{im: rpc, ws: &fakeNotificationSender{}, dlq: &fakeDeadLetterPublisher{}, retryDelay: func(int) time.Duration { return time.Hour }}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := transfer.Consume(ctx, "", validNotificationJSON())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if len(rpc.requests) != 0 {
		t.Fatalf("RPC calls = %d, want 0", len(rpc.requests))
	}
}

func TestCommonNotifyTransferShutdownStopsBackgroundHandlerRetry(t *testing.T) {
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	rpc := &failingNotificationCreator{called: make(chan struct{}, 1)}
	transfer := &CommonNotifyTransfer{
		im: rpc, ws: &fakeNotificationSender{}, dlq: &fakeDeadLetterPublisher{},
		shutdown: shutdownCtx, retryDelay: func(int) time.Duration { return time.Hour },
	}
	done := make(chan error, 1)
	go func() { done <- transfer.Consume(context.Background(), "", validNotificationJSON()) }()
	<-rpc.called

	shutdown()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Consume error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("active retry did not exit after service shutdown")
	}
}

func TestCommonNotifyTransferShutdownCancelsBlockedRPC(t *testing.T) {
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	rpc := &blockingNotificationCreator{called: make(chan struct{})}
	transfer := &CommonNotifyTransfer{
		im: rpc, ws: &fakeNotificationSender{}, dlq: &fakeDeadLetterPublisher{},
		shutdown: shutdownCtx, retryDelay: func(int) time.Duration { return time.Hour },
	}
	done := make(chan error, 1)
	go func() { done <- transfer.Consume(context.Background(), "", validNotificationJSON()) }()
	<-rpc.called

	shutdown()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Consume error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked RPC did not exit after service shutdown")
	}
}

func TestCommonNotifyTransferShutdownCancelsBlockedDLQWrite(t *testing.T) {
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	dlq := &blockingDeadLetterPublisher{called: make(chan struct{})}
	transfer := &CommonNotifyTransfer{
		im: &fakeNotificationCreator{}, ws: &fakeNotificationSender{}, dlq: dlq,
		shutdown: shutdownCtx, retryDelay: func(int) time.Duration { return time.Hour },
	}
	done := make(chan error, 1)
	go func() { done <- transfer.Consume(context.Background(), "", `{`) }()
	<-dlq.called

	shutdown()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Consume error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("blocked DLQ write did not exit after service shutdown")
	}
}

func TestCommonNotifyTransferRejectsPrefetchedMessageAfterShutdown(t *testing.T) {
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	shutdown()
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{{resp: &im.CreateNotificationResp{Inserted: true}}}}
	transfer := &CommonNotifyTransfer{
		im: rpc, ws: &fakeNotificationSender{}, dlq: &fakeDeadLetterPublisher{},
		shutdown: shutdownCtx, retryDelay: func(int) time.Duration { return time.Hour },
	}

	err := transfer.Consume(context.Background(), "", validNotificationJSON())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Consume error = %v, want context.Canceled", err)
	}
}

func TestCommonNotifyTransferAuthenticationErrorsBlockWithoutDLQ(t *testing.T) {
	for _, code := range []codes.Code{codes.Unauthenticated, codes.PermissionDenied} {
		t.Run(code.String(), func(t *testing.T) {
			rpc := &fakeNotificationCreator{results: []notificationRPCResult{{err: status.Error(code, "bad deployment secret")}}}
			dlq := &fakeDeadLetterPublisher{}
			ctx, cancel := context.WithCancel(context.Background())
			transfer := &CommonNotifyTransfer{im: rpc, ws: &fakeNotificationSender{}, dlq: dlq, retryDelay: func(int) time.Duration {
				cancel()
				return time.Hour
			}}

			if err := transfer.Consume(ctx, "", validNotificationJSON()); !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v, want context.Canceled", err)
			}
			if len(rpc.requests) != 1 || len(dlq.values) != 0 {
				t.Fatalf("RPC calls=%d DLQ publishes=%d, want 1 and 0", len(rpc.requests), len(dlq.values))
			}
		})
	}
}

func TestCommonNotifyTransferPoisonToDLQ(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		rpc            *fakeNotificationCreator
		classification string
	}{
		{name: "malformed JSON", raw: `{`, rpc: &fakeNotificationCreator{}, classification: dlqMalformedJSON},
		{name: "RPC invalid argument", raw: validNotificationJSON(), rpc: &fakeNotificationCreator{results: []notificationRPCResult{{err: status.Error(codes.InvalidArgument, "bad receiver")}}}, classification: dlqRPCInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dlq := &fakeDeadLetterPublisher{}
			transfer := newTestNotificationTransfer(tt.rpc, &fakeNotificationSender{}, dlq)
			if err := transfer.Consume(context.Background(), "", tt.raw); err != nil {
				t.Fatal(err)
			}
			if len(dlq.values) != 1 {
				t.Fatalf("DLQ publishes = %d, want 1", len(dlq.values))
			}
			var dead notificationDeadLetter
			if err := json.Unmarshal([]byte(dlq.values[0]), &dead); err != nil {
				t.Fatal(err)
			}
			if dead.Version != 1 || dead.Classification != tt.classification || dead.RawEvent != tt.raw {
				t.Fatalf("dead letter = %+v", dead)
			}
		})
	}
}

func TestCommonNotifyTransferRetriesDLQPublish(t *testing.T) {
	dlq := &fakeDeadLetterPublisher{errors: []error{errors.New("kafka unavailable"), nil}}
	transfer := newTestNotificationTransfer(&fakeNotificationCreator{}, &fakeNotificationSender{}, dlq)

	if err := transfer.Consume(context.Background(), "", `{`); err != nil {
		t.Fatal(err)
	}
	if len(dlq.values) != 2 || dlq.values[0] != dlq.values[1] {
		t.Fatalf("DLQ values = %#v, want two identical attempts", dlq.values)
	}
}

func TestCommonNotifyTransferSuppressesDuplicateAndAlreadyRead(t *testing.T) {
	for _, resp := range []*im.CreateNotificationResp{{Inserted: false}, {Inserted: true, AlreadyRead: true}} {
		rpc := &fakeNotificationCreator{results: []notificationRPCResult{{resp: resp}}}
		sender := &fakeNotificationSender{}
		if err := newTestNotificationTransfer(rpc, sender, &fakeDeadLetterPublisher{}).Consume(context.Background(), "", validNotificationJSON()); err != nil {
			t.Fatal(err)
		}
		if len(sender.messages) != 0 {
			t.Fatalf("websocket sends = %d, want 0 for %+v", len(sender.messages), resp)
		}
	}
}

func TestCommonNotifyTransferRetriesEmptyRPCResponseUntilShutdown(t *testing.T) {
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{{}}}
	transfer := &CommonNotifyTransfer{
		im: rpc, ws: &fakeNotificationSender{}, dlq: &fakeDeadLetterPublisher{},
		shutdown: shutdownCtx, retryDelay: func(int) time.Duration {
			shutdown()
			return time.Hour
		},
	}
	if err := transfer.Consume(context.Background(), "", validNotificationJSON()); !errors.Is(err, context.Canceled) {
		t.Fatalf("Consume error = %v, want context.Canceled", err)
	}
	if len(rpc.requests) != 1 {
		t.Fatalf("RPC calls = %d, want 1", len(rpc.requests))
	}
}

func TestCommonNotifyTransferIgnoresEmptyReceiverAndSelfNotification(t *testing.T) {
	for _, raw := range []string{
		`{"receiverId":"","notifyType":"friend.apply","bizId":"friend:1","actorId":"43"}`,
		`{"receiverId":"43","notifyType":"friend.apply","bizId":"friend:1","actorId":"43"}`,
	} {
		rpc := &fakeNotificationCreator{}
		if err := newTestNotificationTransfer(rpc, &fakeNotificationSender{}, &fakeDeadLetterPublisher{}).Consume(context.Background(), "", raw); err != nil {
			t.Fatal(err)
		}
		if len(rpc.requests) != 0 {
			t.Fatalf("RPC calls = %d, want 0", len(rpc.requests))
		}
	}
}

func TestCommonNotifyTransferDefaultsCreateTimeForRPCAndPush(t *testing.T) {
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{{resp: &im.CreateNotificationResp{Inserted: true}}}}
	sender := &fakeNotificationSender{}
	if err := newTestNotificationTransfer(rpc, sender, &fakeDeadLetterPublisher{}).Consume(context.Background(), "", `{"receiverId":"42","notifyType":"friend.apply","bizId":"friend:1","actorId":"43"}`); err != nil {
		t.Fatal(err)
	}
	if rpc.requests[0].CreateTime == 0 {
		t.Fatal("RPC create time was not defaulted")
	}
	notify := sender.messages[0].(websocket.Message).Data.(*ws.Notify)
	if notify.CreateTime != rpc.requests[0].CreateTime {
		t.Fatalf("push create time = %d, want %d", notify.CreateTime, rpc.requests[0].CreateTime)
	}
}

func TestCommonNotifyTransferWebsocketFailureIsBestEffort(t *testing.T) {
	rpc := &fakeNotificationCreator{results: []notificationRPCResult{{resp: &im.CreateNotificationResp{Inserted: true}}}}
	sender := &fakeNotificationSender{err: errors.New("offline")}
	err := newTestNotificationTransfer(rpc, sender, &fakeDeadLetterPublisher{}).Consume(context.Background(), "", validNotificationJSON())
	if err != nil {
		t.Fatalf("Consume error = %v, want nil", err)
	}
}

func validNotificationJSON() string {
	return `{"receiverId":"42","notifyType":"friend.apply","bizId":"friend:1","actorId":"43","createTime":1710000000}`
}
