package group

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupAPIRecorder struct {
	socialclient.Social
	actors []string
	legacy []string
}

func (r *groupAPIRecorder) MarkGroupReqRead(_ context.Context, in *socialclient.MarkGroupReqReadReq, _ ...grpc.CallOption) (*socialclient.MarkGroupReqReadResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.MarkGroupReqReadResp{}, nil
}

func (r *groupAPIRecorder) GroupInvitationRead(_ context.Context, in *socialclient.GroupInvitationReadReq, _ ...grpc.CallOption) (*socialclient.GroupInvitationReadResp, error) {
	r.record(in.ActorUid, in.ActorUid)
	return &socialclient.GroupInvitationReadResp{}, nil
}

func (r *groupAPIRecorder) GroupRequestMessageCount(_ context.Context, in *socialclient.GroupRequestMessageCountReq, _ ...grpc.CallOption) (*socialclient.GroupRequestMessageCountResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.GroupRequestMessageCountResp{}, nil
}

func (r *groupAPIRecorder) GroupPutinList(_ context.Context, in *socialclient.GroupPutinListReq, _ ...grpc.CallOption) (*socialclient.GroupPutinListResp, error) {
	r.record(in.ActorUid, in.UserId)
	return nil, status.Error(codes.Internal, "stop after recording request")
}

func (r *groupAPIRecorder) GetGroupPutListByUid(_ context.Context, in *socialclient.GetGroupPutListByUidReq, _ ...grpc.CallOption) (*socialclient.GroupPutinListResp, error) {
	legacy := ""
	if len(in.Ids) == 1 {
		legacy = in.Ids[0]
	}
	r.record(in.ActorUid, legacy)
	return &socialclient.GroupPutinListResp{}, nil
}

func (r *groupAPIRecorder) GroupSetAdmin(_ context.Context, in *socialclient.GroupSetAdminReq, _ ...grpc.CallOption) (*socialclient.GroupSetAdminResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.GroupSetAdminResp{}, nil
}

func (r *groupAPIRecorder) GroupInviteLinkCreate(_ context.Context, in *socialclient.GroupInviteLinkCreateReq, _ ...grpc.CallOption) (*socialclient.GroupInviteLinkCreateResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.GroupInviteLinkCreateResp{}, nil
}

func (r *groupAPIRecorder) GroupInviteLinkList(_ context.Context, in *socialclient.GroupInviteLinkListReq, _ ...grpc.CallOption) (*socialclient.GroupInviteLinkListResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.GroupInviteLinkListResp{}, nil
}

func (r *groupAPIRecorder) GroupInviteLinkRevoke(_ context.Context, in *socialclient.GroupInviteLinkRevokeReq, _ ...grpc.CallOption) (*socialclient.GroupInviteLinkRevokeResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.GroupInviteLinkRevokeResp{}, nil
}

func (r *groupAPIRecorder) GroupJoinByToken(_ context.Context, in *socialclient.GroupJoinByTokenReq, _ ...grpc.CallOption) (*socialclient.GroupJoinByTokenResp, error) {
	r.record(in.ActorUid, in.UserId)
	return &socialclient.GroupJoinByTokenResp{}, nil
}

func (r *groupAPIRecorder) record(actor, legacy string) {
	r.actors = append(r.actors, actor)
	r.legacy = append(r.legacy, legacy)
}

func TestGroupReadAndCountBindJWTActor(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxdata.Identify, "9007199254740993")
	recorder := &groupAPIRecorder{}
	svcCtx := &svc.ServiceContext{Social: recorder}

	_, err := NewGroupPutInsReadLogic(ctx, svcCtx).GroupPutInsRead(&types.GroupPutInsReadReq{RequestIds: []string{"18446744073709551615"}})
	require.NoError(t, err)
	_, err = NewGroupInvitationReadLogic(ctx, svcCtx).GroupInvitationRead(&types.GroupInvitationReadReq{InvitationIds: []string{"9007199254740993"}})
	require.NoError(t, err)
	_, err = NewGroupRequestMessageCountLogic(ctx, svcCtx).GroupRequestMessageCount(&types.GroupRequestMessageCountReq{})
	require.NoError(t, err)
	_, err = NewGroupPutInListLogic(ctx, svcCtx).GroupPutInList(&types.GroupPutInListReq{Class: 1})
	require.Equal(t, codes.Internal, status.Code(err))
	_, err = NewGetGroupPutListByUidLogic(ctx, svcCtx).GetGroupPutListByUid(&types.GetGroupPutListByUidReq{Class: "1"})
	require.NoError(t, err)
	_, err = NewGroupSetAdminLogic(ctx, svcCtx).GroupSetAdmin(&types.GroupSetAdminReq{GroupId: "1"})
	require.NoError(t, err)
	_, err = NewGroupInviteLinkCreateLogic(ctx, svcCtx).GroupInviteLinkCreate(&types.GroupInviteLinkCreateReq{GroupId: "1"})
	require.NoError(t, err)
	_, err = NewGroupInviteLinkListLogic(ctx, svcCtx).GroupInviteLinkList(&types.GroupInviteLinkListReq{GroupId: "1"})
	require.NoError(t, err)
	_, err = NewGroupInviteLinkRevokeLogic(ctx, svcCtx).GroupInviteLinkRevoke(&types.GroupInviteLinkRevokeReq{GroupId: "1", Token: "token"})
	require.NoError(t, err)
	_, err = NewGroupJoinByTokenLogic(ctx, svcCtx).GroupJoinByToken(&types.GroupJoinByTokenReq{Token: "token"})
	require.NoError(t, err)

	require.Len(t, recorder.actors, 10)
	for _, actor := range recorder.actors {
		require.Equal(t, "9007199254740993", actor)
	}
	require.Equal(t, recorder.actors, recorder.legacy)
}

func TestGroupReadAndCountRejectInvalidJWTActor(t *testing.T) {
	for _, actor := range []any{nil, "", "invalid", "0", 1} {
		t.Run(fmt.Sprint(actor), func(t *testing.T) {
			ctx := context.Background()
			if actor != nil {
				ctx = context.WithValue(ctx, ctxdata.Identify, actor)
			}
			svcCtx := &svc.ServiceContext{}
			_, err := NewGroupPutInsReadLogic(ctx, svcCtx).GroupPutInsRead(&types.GroupPutInsReadReq{RequestIds: []string{"1"}})
			require.Equal(t, codes.Unauthenticated, status.Code(err))
			_, err = NewGroupInvitationReadLogic(ctx, svcCtx).GroupInvitationRead(&types.GroupInvitationReadReq{InvitationIds: []string{"1"}})
			require.Equal(t, codes.Unauthenticated, status.Code(err))
			_, err = NewGroupRequestMessageCountLogic(ctx, svcCtx).GroupRequestMessageCount(&types.GroupRequestMessageCountReq{})
			require.Equal(t, codes.Unauthenticated, status.Code(err))
			_, err = NewGroupPutInListLogic(ctx, svcCtx).GroupPutInList(&types.GroupPutInListReq{})
			require.Equal(t, codes.Unauthenticated, status.Code(err))
		})
	}
}

func TestGroupInvitationHandleRejectsLongReason(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxdata.Identify, "3")
	_, err := NewGroupInvitationHandleLogic(ctx, &svc.ServiceContext{}).GroupInvitationHandle(&types.GroupInvitationHandleReq{
		Id: "1", Result: 2, HandleMsg: strings.Repeat("拒", 256),
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
