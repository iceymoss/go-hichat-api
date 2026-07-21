package logic

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestNotificationRPCsRequireVerifiedPrincipal(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{name: "list", call: func() error {
			_, err := NewListNotificationsLogic(context.Background(), &svc.ServiceContext{}).ListNotifications(&im.ListNotificationsReq{ReceiverId: "42"})
			return err
		}},
		{name: "unread count", call: func() error {
			_, err := NewGetUnreadNotificationCountLogic(context.Background(), &svc.ServiceContext{}).GetUnreadNotificationCount(&im.GetUnreadNotificationCountReq{ReceiverId: "42"})
			return err
		}},
		{name: "mark read", call: func() error {
			_, err := NewMarkNotificationsReadLogic(context.Background(), &svc.ServiceContext{}).MarkNotificationsRead(&im.MarkNotificationsReadReq{ReceiverId: "42"})
			return err
		}},
		{name: "create", call: func() error {
			_, err := NewCreateNotificationLogic(context.Background(), &svc.ServiceContext{}).CreateNotification(&im.CreateNotificationReq{ReceiverId: "42", NotifyType: "friend.apply", BizId: "friend:1:apply"})
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, codes.Unauthenticated, status.Code(tt.call()))
		})
	}
}

func verifiedUserContext(t *testing.T, uid, method string, request proto.Message) (context.Context, *rpcauth.Auth) {
	t.Helper()
	auth, err := rpcauth.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	outgoing, err := rpcauth.WithUser(context.Background(), uid)
	require.NoError(t, err)
	var signed context.Context
	err = auth.UnaryClientInterceptor()(outgoing, method, request, nil, nil,
		func(ctx context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			signed = ctx
			return nil
		})
	require.NoError(t, err)
	md, ok := metadata.FromOutgoingContext(signed)
	require.True(t, ok)
	incoming := metadata.NewIncomingContext(context.Background(), md)

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	var verified context.Context
	_, err = auth.UnaryServerInterceptor(rpcauth.NewRedisReplayStore(redisClient), method)(
		incoming, request, &grpc.UnaryServerInfo{FullMethod: method},
		func(ctx context.Context, _ interface{}) (interface{}, error) {
			verified = ctx
			return nil, nil
		})
	require.NoError(t, err)
	return verified, auth
}

func TestListNotificationsPaginationValidation(t *testing.T) {
	for _, values := range [][2]int32{
		{-1, 20},
		{100001, 20},
		{0, -1},
		{0, 101},
	} {
		require.Equal(t, codes.InvalidArgument, status.Code(validateNotificationPagination(values[0], values[1])))
	}
	require.NoError(t, validateNotificationPagination(100000, 100))
}

func TestNotificationUserRPCsRejectNonCanonicalReceiver(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	_, err := NewListNotificationsLogic(context.Background(), svcCtx).ListNotifications(&im.ListNotificationsReq{ReceiverId: "042"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = NewGetUnreadNotificationCountLogic(context.Background(), svcCtx).GetUnreadNotificationCount(&im.GetUnreadNotificationCountReq{ReceiverId: "042"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	_, err = NewMarkNotificationsReadLogic(context.Background(), svcCtx).MarkNotificationsRead(&im.MarkNotificationsReadReq{ReceiverId: "042"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateNotificationInputBounds(t *testing.T) {
	now := time.Unix(1_750_000_000, 0)
	for _, req := range []*im.CreateNotificationReq{
		nil,
		{ReceiverId: "042", NotifyType: "friend.apply", BizId: "friend:1:apply"},
		{ReceiverId: "42", BizId: "friend:1:apply"},
		{ReceiverId: "42", NotifyType: "friend.apply"},
		{ReceiverId: "42", NotifyType: strings.Repeat("x", 65), BizId: "biz"},
		{ReceiverId: "42", NotifyType: "type", BizId: "biz", Payload: strings.Repeat("x", 65536)},
		{ReceiverId: "42", NotifyType: "type", BizId: "biz", CreateTime: now.Add(-30*time.Hour*24 - time.Second).Unix()},
		{ReceiverId: "42", NotifyType: "type", BizId: "biz", CreateTime: now.Add(5*time.Minute + time.Second).Unix()},
	} {
		require.Equal(t, codes.InvalidArgument, status.Code(validateCreateNotification(req, now)))
	}
	require.NoError(t, validateCreateNotification(&im.CreateNotificationReq{
		ReceiverId: "42", NotifyType: "type", BizId: "biz", Payload: strings.Repeat("x", 65535), CreateTime: now.Unix(),
	}, now))
}
