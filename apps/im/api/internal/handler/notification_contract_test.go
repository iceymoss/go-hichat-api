package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestMarkNotificationsReadDecodesSnakeCaseTargets(t *testing.T) {
	req := httptest.NewRequest("POST", "/v1/im/notifications/read", strings.NewReader(`{
		"targets":[{"notify_type":"group.removed","biz_id":"group.removed:7:9:1"}]
	}`))
	req.Header.Set("Content-Type", "application/json")

	var decoded types.MarkNotificationsReadReq
	require.NoError(t, httpx.Parse(req, &decoded))
	require.Equal(t, []types.NotificationReadTarget{{
		NotifyType: "group.removed",
		BizId:      "group.removed:7:9:1",
	}}, decoded.Targets)
}
