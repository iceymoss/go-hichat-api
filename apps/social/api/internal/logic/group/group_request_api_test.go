package group

import (
	"context"
	"fmt"
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
	markReadActor   string
	invitationActor string
	countActor      string
}

func (r *groupAPIRecorder) MarkGroupReqRead(_ context.Context, in *socialclient.MarkGroupReqReadReq, _ ...grpc.CallOption) (*socialclient.MarkGroupReqReadResp, error) {
	r.markReadActor = in.UserId
	return &socialclient.MarkGroupReqReadResp{}, nil
}

func (r *groupAPIRecorder) GroupInvitationRead(_ context.Context, in *socialclient.GroupInvitationReadReq, _ ...grpc.CallOption) (*socialclient.GroupInvitationReadResp, error) {
	r.invitationActor = in.ActorUid
	return &socialclient.GroupInvitationReadResp{}, nil
}

func (r *groupAPIRecorder) GroupRequestMessageCount(_ context.Context, in *socialclient.GroupRequestMessageCountReq, _ ...grpc.CallOption) (*socialclient.GroupRequestMessageCountResp, error) {
	r.countActor = in.UserId
	return &socialclient.GroupRequestMessageCountResp{}, nil
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

	require.Equal(t, "9007199254740993", recorder.markReadActor)
	require.Equal(t, "9007199254740993", recorder.invitationActor)
	require.Equal(t, "9007199254740993", recorder.countActor)
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
