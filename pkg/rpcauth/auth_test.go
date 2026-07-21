package rpcauth_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/apps/im/rpc/im"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const testSecret = "0123456789abcdef0123456789abcdef"

type unavailableReplayStore struct{}

func (unavailableReplayStore) Consume(context.Context, string, time.Duration) (bool, error) {
	return false, errors.Join(rpcauth.ErrReplayStore, errors.New("redis unavailable"))
}

type authTestServer struct {
	im.UnimplementedImServer
}

func (*authTestServer) ListNotifications(ctx context.Context, in *im.ListNotificationsReq) (*im.ListNotificationsResp, error) {
	if err := rpcauth.RequireUser(ctx, in.ReceiverId); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &im.ListNotificationsResp{}, nil
}

func (*authTestServer) CreateNotification(ctx context.Context, _ *im.CreateNotificationReq) (*im.CreateNotificationResp, error) {
	if err := rpcauth.RequireTask(ctx); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &im.CreateNotificationResp{}, nil
}

func TestUnaryAuthenticationBindsMethodBodyAndNonce(t *testing.T) {
	auth, err := rpcauth.New(testSecret)
	require.NoError(t, err)
	client, conn, captured := newAuthenticatedClient(t, auth, auth, nil, nil)
	defer conn.Close()

	ctx, err := rpcauth.WithUser(context.Background(), "42")
	require.NoError(t, err)
	_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42", Limit: 20})
	require.NoError(t, err)
	firstNonce := captured().Get("x-hichat-rpc-nonce")
	require.Len(t, firstNonce, 1)

	ctx, err = rpcauth.WithUser(context.Background(), "42")
	require.NoError(t, err)
	_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42", Limit: 20})
	require.NoError(t, err)
	secondNonce := captured().Get("x-hichat-rpc-nonce")
	require.Len(t, secondNonce, 1)
	require.NotEqual(t, firstNonce[0], secondNonce[0])

	replayedMetadata := captured()
	replayed := metadata.NewOutgoingContext(context.Background(), replayedMetadata)
	_, err = im.NewImClient(conn).ListNotifications(replayed, &im.ListNotificationsReq{ReceiverId: "42", Limit: 20})
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	t.Run("duplicate metadata", func(t *testing.T) {
		duplicate := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			md, _ := metadata.FromOutgoingContext(ctx)
			md = md.Copy()
			md.Append("x-hichat-rpc-signature", md.Get("x-hichat-rpc-signature")[0])
			return invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...)
		}
		client, conn, _ := newAuthenticatedClient(t, auth, auth, duplicate, nil)
		defer conn.Close()
		ctx, err := rpcauth.WithUser(context.Background(), "42")
		require.NoError(t, err)
		_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42"})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("body substitution", func(t *testing.T) {
		tamper := func(ctx context.Context, method string, _, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, &im.ListNotificationsReq{ReceiverId: "43", Limit: 20}, reply, cc, opts...)
		}
		client, conn, _ := newAuthenticatedClient(t, auth, auth, tamper, nil)
		defer conn.Close()
		ctx, err := rpcauth.WithUser(context.Background(), "42")
		require.NoError(t, err)
		_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42", Limit: 20})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("method substitution", func(t *testing.T) {
		tamper := func(ctx context.Context, _ string, _ interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, im.Im_CreateNotification_FullMethodName, &im.CreateNotificationReq{ReceiverId: "42"}, reply, cc, opts...)
		}
		client, conn, _ := newAuthenticatedClient(t, auth, auth, tamper, nil)
		defer conn.Close()
		ctx, err := rpcauth.WithUser(context.Background(), "42")
		require.NoError(t, err)
		_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42"})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestUnaryAuthenticationRejectsWrongSecretAndFailsClosed(t *testing.T) {
	serverAuth, err := rpcauth.New(testSecret)
	require.NoError(t, err)
	otherAuth, err := rpcauth.New("abcdef0123456789abcdef0123456789")
	require.NoError(t, err)

	t.Run("wrong secret", func(t *testing.T) {
		client, conn, _ := newAuthenticatedClient(t, otherAuth, serverAuth, nil, nil)
		defer conn.Close()
		ctx, err := rpcauth.WithUser(context.Background(), "42")
		require.NoError(t, err)
		_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42"})
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("replay store unavailable", func(t *testing.T) {
		client, conn, _ := newAuthenticatedClient(t, serverAuth, serverAuth, nil, unavailableReplayStore{})
		defer conn.Close()
		ctx, err := rpcauth.WithUser(context.Background(), "42")
		require.NoError(t, err)
		_, err = client.ListNotifications(ctx, &im.ListNotificationsReq{ReceiverId: "42"})
		require.Equal(t, codes.Unavailable, status.Code(err))
	})
}

func TestCreateNotificationAllowsOnlyTaskPrincipal(t *testing.T) {
	auth, err := rpcauth.New(testSecret)
	require.NoError(t, err)
	client, conn, _ := newAuthenticatedClient(t, auth, auth, nil, nil)
	defer conn.Close()

	_, err = client.CreateNotification(rpcauth.WithTask(context.Background()), &im.CreateNotificationReq{ReceiverId: "42"})
	require.NoError(t, err)

	userCtx, err := rpcauth.WithUser(context.Background(), "42")
	require.NoError(t, err)
	_, err = client.CreateNotification(userCtx, &im.CreateNotificationReq{ReceiverId: "42"})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestSecretLoadingAgreement(t *testing.T) {
	t.Setenv(rpcauth.EnvSecret, testSecret)
	require.Equal(t, testSecret, rpcauth.LoadSecret("different-config-value"))
	_, err := rpcauth.New(rpcauth.LoadSecret(""))
	require.NoError(t, err)

	old, present := os.LookupEnv(rpcauth.EnvSecret)
	require.True(t, present)
	require.Equal(t, testSecret, old)
}

func TestEmptyEnvironmentUsesConfiguredDevelopmentSecret(t *testing.T) {
	t.Setenv(rpcauth.EnvSecret, "")
	require.Equal(t, testSecret, rpcauth.LoadSecret(testSecret))
}

func TestAPIAndRPCLoadSameDedicatedEnvironmentSecret(t *testing.T) {
	t.Setenv(rpcauth.EnvSecret, testSecret)
	apiSecret := rpcauth.LoadSecret("")
	rpcSecret := rpcauth.LoadSecret("")
	require.Equal(t, apiSecret, rpcSecret)
	_, err := rpcauth.New(apiSecret)
	require.NoError(t, err)
}

func TestNewRequiresStrongDedicatedSecret(t *testing.T) {
	for _, secret := range []string{"", "short-secret"} {
		auth, err := rpcauth.New(secret)
		require.Nil(t, auth)
		require.ErrorIs(t, err, rpcauth.ErrMissingSecret)
	}
}

func TestCanonicalUID(t *testing.T) {
	for _, uid := range []string{"", "0", "01", "-1", "user", "18446744073709551616"} {
		require.False(t, rpcauth.CanonicalUID(uid), uid)
	}
	require.True(t, rpcauth.CanonicalUID("1"))
}

func newAuthenticatedClient(
	t *testing.T,
	clientAuth, serverAuth *rpcauth.Auth,
	tamper grpc.UnaryClientInterceptor,
	store rpcauth.ReplayStore,
) (im.ImClient, *grpc.ClientConn, func() metadata.MD) {
	t.Helper()
	if store == nil {
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
		t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
		store = rpcauth.NewRedisReplayStore(redisClient)
	}

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(serverAuth.UnaryServerInterceptor(
		store,
		im.Im_CreateNotification_FullMethodName,
		im.Im_ListNotifications_FullMethodName,
	)))
	im.RegisterImServer(server, &authTestServer{})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		require.NoError(t, listener.Close())
	})

	var signedMetadata metadata.MD
	capture := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		signedMetadata, _ = metadata.FromOutgoingContext(ctx)
		signedMetadata = signedMetadata.Copy()
		if tamper != nil {
			return tamper(ctx, method, req, reply, cc, invoker, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientAuth.UnaryClientInterceptor(), capture),
	)
	require.NoError(t, err)
	return im.NewImClient(conn), conn, func() metadata.MD { return signedMetadata.Copy() }
}
