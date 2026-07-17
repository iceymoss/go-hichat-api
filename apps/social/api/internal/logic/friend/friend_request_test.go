package friend

import (
	"context"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
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
