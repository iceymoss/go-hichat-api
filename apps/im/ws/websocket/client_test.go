package websocket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestClosedWsClientDoesNotReconnect(t *testing.T) {
	client := &client{host: "127.0.0.1:1"}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := client.Send(struct{}{}); !errors.Is(err, errClientClosed) {
		t.Fatalf("Send error = %v, want errClientClosed", err)
	}
}

func TestWsClientCloseCancelsActiveDial(t *testing.T) {
	dialCtx, cancelDial := context.WithCancel(context.Background())
	started := make(chan struct{})
	client := &client{
		host:       "example.invalid",
		dialCtx:    dialCtx,
		cancelDial: cancelDial,
		dialContext: func(ctx context.Context, _ string, _ http.Header) (*websocket.Conn, *http.Response, error) {
			close(started)
			<-ctx.Done()
			return nil, nil, ctx.Err()
		},
	}
	done := make(chan error, 1)
	go func() {
		_, err := client.dail()
		done <- err
	}()
	<-started
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("dial error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("active dial did not exit after Close")
	}
}

func TestWsClientHeartbeatStopsOnClose(t *testing.T) {
	received := make(chan struct{}, 2)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			var message Message
			if err := conn.ReadJSON(&message); err != nil {
				return
			}
			if message.FrameType == FrameNoAck && message.Method == "chat.ping" {
				received <- struct{}{}
			}
		}
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	client := NewClient(host, WithClientPatten("/"), withClientHeartbeat(10*time.Millisecond))
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("heartbeat was not sent")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-received:
		t.Fatal("heartbeat continued after Close")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWsClient(t *testing.T) {
	host := os.Getenv("HICHAT_WS_TEST_HOST")
	if host == "" {
		t.Skip("set HICHAT_WS_TEST_HOST to run the external websocket integration test")
	}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njc0MTMzODAsImhpY2hhdDIuY29tIjoiMTgiLCJpYXQiOjE3NTg3NzMzODB9.ri6328JcnevAsS3FjEwpOwQKcIwxly5gnNtsO_6acQo"
	header := http.Header{}
	header.Set("Authorization", token)
	wsClient := NewClient(host, WithClientHeader(header))
	if wsClient == nil {
		t.Error("wsClient is nil")
		return
	}
}
