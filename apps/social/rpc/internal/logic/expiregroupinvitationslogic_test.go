package logic

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/db/objects"
	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"

	"github.com/alicebob/miniredis/v2"
	redisv8 "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestExpireGroupInvitationsBatchesReceiptsAndIsIdempotent(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	now := time.Now()
	for i := 0; i < 3; i++ {
		invitation := objects.GroupInvitation{
			GroupID: 1, InviterUID: 1, InviteeUID: uint64(i + 2), Status: groupInvitationPending,
			CreatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-time.Minute),
		}
		require.NoError(t, svcCtx.DB.Create(&invitation).Error)
		require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroupInvite, invitation.ID, fmt.Sprint(i+2), receiptKindInvite, 1, receiptPending, invitation.CreatedAt, false, nil))
	}
	future := objects.GroupInvitation{GroupID: 1, InviterUID: 1, InviteeUID: 9, Status: groupInvitationPending, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	require.NoError(t, svcCtx.DB.Create(&future).Error)

	logic := NewExpireGroupInvitationsLogic(verifiedExpirationTaskContext(t), svcCtx)
	first, err := logic.ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{BatchSize: 2})
	require.NoError(t, err)
	require.Equal(t, int32(2), first.Expired)
	require.True(t, first.HasMore)
	second, err := logic.ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{BatchSize: 2})
	require.NoError(t, err)
	require.Equal(t, int32(1), second.Expired)
	require.False(t, second.HasMore)
	third, err := logic.ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{BatchSize: 2})
	require.NoError(t, err)
	require.Zero(t, third.Expired)

	var receipts []objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Order("request_id ASC").Find(&receipts).Error)
	require.Len(t, receipts, 3)
	for _, receipt := range receipts {
		require.Equal(t, receiptExpired, receipt.Result)
		require.Zero(t, receipt.IsActionable)
		require.NotNil(t, receipt.ResolvedAt)
	}
	var pending int64
	require.NoError(t, svcCtx.DB.Model(&objects.GroupInvitation{}).Where("status = ?", groupInvitationPending).Count(&pending).Error)
	require.Equal(t, int64(1), pending)
}

func TestExpireGroupInvitationsReceiptFailureRollsBack(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger injects receipt failure")
	}
	now := time.Now()
	invitation := objects.GroupInvitation{GroupID: 1, InviterUID: 1, InviteeUID: 2, Status: groupInvitationPending, CreatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-time.Minute)}
	require.NoError(t, svcCtx.DB.Create(&invitation).Error)
	require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroupInvite, invitation.ID, "2", receiptKindInvite, 1, receiptPending, invitation.CreatedAt, false, nil))
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER fail_expire_receipt BEFORE UPDATE ON social_request_receipts BEGIN SELECT RAISE(ABORT, 'receipt failure'); END`).Error)

	_, err := NewExpireGroupInvitationsLogic(verifiedExpirationTaskContext(t), svcCtx).ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{BatchSize: 1})
	require.Equal(t, codes.Internal, status.Code(err))
	var latest objects.GroupInvitation
	require.NoError(t, svcCtx.DB.First(&latest, invitation.ID).Error)
	require.Equal(t, groupInvitationPending, latest.Status)
}

func TestExpireGroupInvitationsConcurrentConfirmationConverges(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	require.NoError(t, svcCtx.DB.Create(testGroup(1, true)).Error)
	seedGroupMember(t, svcCtx.DB, 1, 1, 2)
	now := time.Now()
	invitation := objects.GroupInvitation{GroupID: 1, InviterUID: 1, InviteeUID: 3, Status: groupInvitationPending, CreatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-time.Minute)}
	require.NoError(t, svcCtx.DB.Create(&invitation).Error)
	require.NoError(t, createReceipt(svcCtx.DB, receiptTypeGroupInvite, invitation.ID, "3", receiptKindInvite, 1, receiptPending, invitation.CreatedAt, false, nil))

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := NewExpireGroupInvitationsLogic(verifiedExpirationTaskContext(t), svcCtx).ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{BatchSize: 1})
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := NewGroupInvitationHandleLogic(context.Background(), svcCtx).GroupInvitationHandle(&social.GroupInvitationHandleReq{Id: invitation.ID, ActorUid: "3", Result: 2})
		errs <- err
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var latest objects.GroupInvitation
	require.NoError(t, svcCtx.DB.First(&latest, invitation.ID).Error)
	require.Equal(t, groupInvitationExpired, latest.Status)
	var receipt objects.SocialRequestReceipt
	require.NoError(t, svcCtx.DB.Where("request_type = ? AND request_id = ?", receiptTypeGroupInvite, invitation.ID).First(&receipt).Error)
	require.Equal(t, receiptExpired, receipt.Result)
	require.Zero(t, receipt.IsActionable)
}

func TestExpireGroupInvitationsRequiresTaskPrincipal(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	_, err := NewExpireGroupInvitationsLogic(context.Background(), svcCtx).ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{})
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	auth, err := rpcauth.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	userCtx, err := rpcauth.WithUser(context.Background(), "42")
	require.NoError(t, err)
	verified := verifyExpirationContext(t, auth, userCtx, &social.ExpireGroupInvitationsReq{})
	_, err = NewExpireGroupInvitationsLogic(verified, svcCtx).ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{})
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	_, err = NewExpireGroupInvitationsLogic(verifiedExpirationTaskContext(t), svcCtx).ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{})
	require.NoError(t, err)
}

func TestExpireGroupInvitationsHasMoreUsesSelectedCandidates(t *testing.T) {
	svcCtx, _ := newGroupTestContext(t)
	if svcCtx.DB.Dialector.Name() != "sqlite" {
		t.Skip("SQLite trigger injects a CAS miss")
	}
	now := time.Now()
	invitation := objects.GroupInvitation{GroupID: 1, InviterUID: 1, InviteeUID: 2, Status: groupInvitationPending, CreatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-time.Minute)}
	require.NoError(t, svcCtx.DB.Create(&invitation).Error)
	require.NoError(t, svcCtx.DB.Exec(`CREATE TRIGGER miss_expiration_cas BEFORE UPDATE OF status ON group_invitations WHEN NEW.status = 4 BEGIN UPDATE group_invitations SET status = 2 WHERE id = OLD.id; SELECT RAISE(IGNORE); END`).Error)

	resp, err := NewExpireGroupInvitationsLogic(verifiedExpirationTaskContext(t), svcCtx).ExpireGroupInvitations(&social.ExpireGroupInvitationsReq{BatchSize: 1})
	require.NoError(t, err)
	require.Zero(t, resp.Expired)
	require.True(t, resp.HasMore)
}

func TestExpireGroupInvitationsServerAuthPrincipal(t *testing.T) {
	auth, err := rpcauth.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	request := &social.ExpireGroupInvitationsReq{BatchSize: 10}
	method := social.Social_ExpireGroupInvitations_FullMethodName
	redisServer := miniredis.RunT(t)
	redisClient := redisv8.NewClient(&redisv8.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	interceptor := auth.UnaryServerInterceptor(rpcauth.NewRedisReplayStore(redisClient), method)

	_, err = interceptor(context.Background(), request, &grpc.UnaryServerInfo{FullMethod: method}, func(context.Context, interface{}) (interface{}, error) {
		return nil, nil
	})
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	for _, tt := range []struct {
		name     string
		outbound context.Context
		wantErr  error
	}{
		{name: "wrong user principal", outbound: mustUserContext(t, "42"), wantErr: rpcauth.ErrWrongPrincipal},
		{name: "valid task principal", outbound: rpcauth.WithTask(context.Background())},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var signed context.Context
			err := auth.UnaryClientInterceptor()(tt.outbound, method, request, nil, nil,
				func(ctx context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
					signed = ctx
					return nil
				})
			require.NoError(t, err)
			md, ok := metadata.FromOutgoingContext(signed)
			require.True(t, ok)
			_, err = interceptor(metadata.NewIncomingContext(context.Background(), md), request, &grpc.UnaryServerInfo{FullMethod: method},
				func(ctx context.Context, _ interface{}) (interface{}, error) {
					return nil, rpcauth.RequireTask(ctx)
				})
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func mustUserContext(t *testing.T, uid string) context.Context {
	t.Helper()
	ctx, err := rpcauth.WithUser(context.Background(), uid)
	require.NoError(t, err)
	return ctx
}

func verifiedExpirationTaskContext(t *testing.T) context.Context {
	t.Helper()
	auth, err := rpcauth.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	return verifyExpirationContext(t, auth, rpcauth.WithTask(context.Background()), &social.ExpireGroupInvitationsReq{})
}

func verifyExpirationContext(t *testing.T, auth *rpcauth.Auth, outbound context.Context, request *social.ExpireGroupInvitationsReq) context.Context {
	t.Helper()
	method := social.Social_ExpireGroupInvitations_FullMethodName
	var signed context.Context
	err := auth.UnaryClientInterceptor()(outbound, method, request, nil, nil,
		func(ctx context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			signed = ctx
			return nil
		})
	require.NoError(t, err)
	md, ok := metadata.FromOutgoingContext(signed)
	require.True(t, ok)
	redisServer := miniredis.RunT(t)
	redisClient := redisv8.NewClient(&redisv8.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	var verified context.Context
	_, err = auth.UnaryServerInterceptor(rpcauth.NewRedisReplayStore(redisClient), method)(
		metadata.NewIncomingContext(context.Background(), md), request, &grpc.UnaryServerInfo{FullMethod: method},
		func(ctx context.Context, _ interface{}) (interface{}, error) {
			verified = ctx
			return nil, nil
		})
	require.NoError(t, err)
	return verified
}
