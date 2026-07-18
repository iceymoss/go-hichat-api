package websocket

import (
	"net/http"
	"os"
	"testing"
)

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
