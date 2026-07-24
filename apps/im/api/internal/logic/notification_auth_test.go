package logic

import (
	"context"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNotificationAPILogicRequiresJWTIdentity(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		logic := NewListNotificationsLogic(context.Background(), &svc.ServiceContext{})
		_, err := logic.ListNotifications(&types.ListNotificationsReq{})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})
	t.Run("mark", func(t *testing.T) {
		logic := NewMarkNotificationsReadLogic(context.Background(), &svc.ServiceContext{})
		_, err := logic.MarkNotificationsRead(&types.MarkNotificationsReadReq{})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})
	t.Run("unread count", func(t *testing.T) {
		logic := NewGetNotificationUnreadCountLogic(context.Background(), &svc.ServiceContext{})
		_, err := logic.GetNotificationUnreadCount()
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}
