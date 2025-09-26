package websocket

import (
	"net/http"
	"testing"
)

func TestWsClient(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Njc0MTMzODAsImhpY2hhdDIuY29tIjoiMTgiLCJpYXQiOjE3NTg3NzMzODB9.ri6328JcnevAsS3FjEwpOwQKcIwxly5gnNtsO_6acQo"
	header := http.Header{}
	header.Set("Authorization", token)
	wsClient := NewClient("117.72.163.141:10090", WithClientHeader(header))
	if wsClient == nil {
		t.Error("wsClient is nil")
		return
	}
}
