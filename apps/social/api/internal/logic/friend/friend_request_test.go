package friend

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiUserLookupStub struct {
	response *user.GetUserByIdResponse
}

type friendAPIRecorder struct {
	socialclient.Social
	actors []string
	legacy []string
}

func (r *friendAPIRecorder) FriendPutInList(_ context.Context, in *socialclient.FriendPutInListReq, _ ...grpc.CallOption) (*socialclient.FriendPutInListResp, error) {
	r.actors = append(r.actors, in.ActorUid)
	r.legacy = append(r.legacy, in.UserId)
	return &socialclient.FriendPutInListResp{}, nil
}

func (r *friendAPIRecorder) FriendPutInRead(_ context.Context, in *socialclient.FriendPutInReadReq, _ ...grpc.CallOption) (*socialclient.FriendPutInReadResp, error) {
	r.actors = append(r.actors, in.ActorUid)
	r.legacy = append(r.legacy, in.UserId)
	return &socialclient.FriendPutInReadResp{}, nil
}

func (r *friendAPIRecorder) FriendPutInMessageCount(_ context.Context, in *socialclient.FriendPutInMessageCountReq, _ ...grpc.CallOption) (*socialclient.FriendPutInMessageCountResp, error) {
	r.actors = append(r.actors, in.ActorUid)
	r.legacy = append(r.legacy, in.UserId)
	return &socialclient.FriendPutInMessageCountResp{}, nil
}

func (r *friendAPIRecorder) FriendPutInDelete(_ context.Context, in *socialclient.FriendPutInDeleteReq, _ ...grpc.CallOption) (*socialclient.FriendPutInDeleteResp, error) {
	r.actors = append(r.actors, in.ActorUid)
	r.legacy = append(r.legacy, in.UserId)
	return &socialclient.FriendPutInDeleteResp{}, nil
}

func (s *apiUserLookupStub) GetUserById(context.Context, *user.GetUserByIdRequest, ...grpc.CallOption) (*user.GetUserByIdResponse, error) {
	return s.response, nil
}

func (s *apiUserLookupStub) FindUser(context.Context, *user.FindUserReq, ...grpc.CallOption) (*user.FindUserResp, error) {
	return nil, nil
}

func TestFriendPutInAPIValidation(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		target string
		user   *user.GetUserByIdResponse
		code   codes.Code
	}{
		{name: "missing identity", ctx: context.Background(), target: "2", code: codes.Unauthenticated},
		{name: "malformed identity", ctx: context.WithValue(context.Background(), ctxdata.Identify, "x"), target: "2", code: codes.Unauthenticated},
		{name: "invalid target", ctx: context.WithValue(context.Background(), ctxdata.Identify, "1"), target: "x", code: codes.InvalidArgument},
		{name: "self", ctx: context.WithValue(context.Background(), ctxdata.Identify, "1"), target: "1", code: codes.InvalidArgument},
		{name: "nil user response", ctx: context.WithValue(context.Background(), ctxdata.Identify, "1"), target: "2", code: codes.NotFound},
		{name: "disabled target", ctx: context.WithValue(context.Background(), ctxdata.Identify, "1"), target: "2", user: &user.GetUserByIdResponse{User: &user.UserEntity{Id: "2", Status: 0}}, code: codes.NotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logic := NewFriendPutInLogic(tt.ctx, &svc.ServiceContext{User: &apiUserLookupStub{response: tt.user}})
			_, err := logic.FriendPutIn(&types.FriendPutInReq{UserId: tt.target})
			require.Equal(t, tt.code, status.Code(err))
		})
	}
}

func TestFriendPutInHandleAPIValidation(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		id     int32
		result int32
		code   codes.Code
	}{
		{name: "missing identity", ctx: context.Background(), id: 1, result: 1, code: codes.Unauthenticated},
		{name: "invalid request id", ctx: context.WithValue(context.Background(), ctxdata.Identify, "2"), result: 1, code: codes.InvalidArgument},
		{name: "invalid result", ctx: context.WithValue(context.Background(), ctxdata.Identify, "2"), id: 1, result: 3, code: codes.InvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logic := NewFriendPutInHandleLogic(tt.ctx, &svc.ServiceContext{})
			_, err := logic.FriendPutInHandle(&types.FriendPutInHandleReq{FriendReqId: tt.id, HandleResult: tt.result})
			require.Equal(t, tt.code, status.Code(err))
		})
	}
}

func TestFriendScopedAPIBindsJWTActor(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxdata.Identify, "9007199254740993")
	recorder := &friendAPIRecorder{}
	svcCtx := &svc.ServiceContext{Social: recorder}

	_, err := NewFriendPutInListLogic(ctx, svcCtx, httptest.NewRequest("GET", "/", nil)).FriendPutInList(&types.FriendPutInListReq{Class: "1"})
	require.NoError(t, err)
	_, err = NewFriendPutInReadLogic(ctx, svcCtx).FriendPutInRead(&types.FriendPutInReadReq{})
	require.NoError(t, err)
	_, err = NewFriendPutInMessageCountLogic(ctx, svcCtx).FriendPutInMessageCount(&types.FriendPutInMessageCountReq{})
	require.NoError(t, err)
	_, err = NewFriendPutInDeleteLogic(ctx, svcCtx).FriendPutInDelete(&types.FriendPutInDeleteReq{RequestId: "1"})
	require.NoError(t, err)

	require.Equal(t, []string{"9007199254740993", "9007199254740993", "9007199254740993", "9007199254740993"}, recorder.actors)
	require.Equal(t, recorder.actors, recorder.legacy)
}

func TestFriendScopedAPIRejectsInvalidJWTActor(t *testing.T) {
	for _, actor := range []any{nil, "", "invalid", "0", "01", 1} {
		t.Run(fmt.Sprint(actor), func(t *testing.T) {
			ctx := context.Background()
			if actor != nil {
				ctx = context.WithValue(ctx, ctxdata.Identify, actor)
			}
			_, err := NewFriendPutInReadLogic(ctx, &svc.ServiceContext{}).FriendPutInRead(&types.FriendPutInReadReq{})
			require.Equal(t, codes.Unauthenticated, status.Code(err))
		})
	}
}
