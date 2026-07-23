package friend

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestFriendPutInHandleRequestParsing(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "request id", body: `{"request_id":"1","handle_result":1}`},
		{name: "legacy friend request id", body: `{"friend_req_id":1,"handle_result":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("PUT", "/api/social/friend/putIn", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			var decoded types.FriendPutInHandleReq
			require.NoError(t, httpx.Parse(req, &decoded))
			require.Equal(t, int32(1), decoded.HandleResult)
		})
	}
}
