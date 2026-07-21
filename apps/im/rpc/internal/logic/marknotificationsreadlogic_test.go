package logic

import (
	"context"
	"strconv"
	"testing"

	models "github.com/iceymoss/go-hichat-api/apps/im/models"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMarkNotificationsReadValidation(t *testing.T) {
	t.Run("receiver required", func(t *testing.T) {
		logic := NewMarkNotificationsReadLogic(context.Background(), &svc.ServiceContext{})
		_, err := logic.MarkNotificationsRead(&im.MarkNotificationsReadReq{ReceiverId: "not-a-user"})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("rejects more than one hundred", func(t *testing.T) {
		targets := make([]*im.NotificationReadTarget, 101)
		for i := range targets {
			targets[i] = &im.NotificationReadTarget{NotifyType: "friend.apply", BizId: "friend:" + strconv.Itoa(i+1) + ":apply"}
		}
		request := &im.MarkNotificationsReadReq{ReceiverId: "1", Targets: targets}
		ctx, auth := verifiedUserContext(t, "1", im.Im_MarkNotificationsRead_FullMethodName, request)
		logic := NewMarkNotificationsReadLogic(ctx, &svc.ServiceContext{RPCAuth: auth})
		_, err := logic.MarkNotificationsRead(request)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestLegacyNotificationReadTargetsAcceptArbitraryExactPairs(t *testing.T) {
	targets, err := legacyNotificationReadTargets(
		[]string{"group.removed", "group.admin.set"},
		[]string{"group.removed:7:9:1", "custom-admin-biz"},
	)
	require.NoError(t, err)
	require.Equal(t, []models.NotificationReadTarget{
		{NotifyType: "group.removed", BizId: "group.removed:7:9:1"},
		{NotifyType: "group.admin.set", BizId: "custom-admin-biz"},
	}, targets)

	_, err = legacyNotificationReadTargets([]string{"group.removed"}, []string{""})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
