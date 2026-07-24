package group

import (
	"context"
	"testing"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type groupDetailSocialStub struct {
	socialclient.Social
	response *socialclient.GroupDetailResp
}

func (s *groupDetailSocialStub) GroupDetail(context.Context, *socialclient.GroupDetailReq, ...grpc.CallOption) (*socialclient.GroupDetailResp, error) {
	return s.response, nil
}

type groupDetailUserStub struct {
	users []*user.UserEntity
}

func (s *groupDetailUserStub) GetUserById(context.Context, *user.GetUserByIdRequest, ...grpc.CallOption) (*user.GetUserByIdResponse, error) {
	return nil, nil
}

func (s *groupDetailUserStub) FindUser(context.Context, *user.FindUserReq, ...grpc.CallOption) (*user.FindUserResp, error) {
	return &user.FindUserResp{User: s.users}, nil
}

func TestGroupDetailToleratesMissingUserProfiles(t *testing.T) {
	logic := NewGroupDetailLogic(context.Background(), &svc.ServiceContext{
		Social: &groupDetailSocialStub{response: &socialclient.GroupDetailResp{
			Group: &socialclient.Groups{Id: "1", Name: "group"},
			Members: []*socialclient.GroupMembers{
				nil,
				{Id: 1, GroupId: "1", UserId: "missing", RoleLevel: 1},
				{Id: 2, GroupId: "1", UserId: "present"},
			},
		}},
		User: &groupDetailUserStub{users: []*user.UserEntity{nil, {Id: "present", Nickname: "Present"}}},
	})

	resp, err := logic.GroupDetail(&types.GroupDetailReq{GroupId: "1"})
	require.NoError(t, err)
	require.Len(t, resp.Members, 2)
	require.Equal(t, "missing", resp.Members[0].UserId)
	require.Equal(t, "missing", resp.Members[0].User.Id)
	require.Empty(t, resp.Members[0].Nickname)
	require.Equal(t, "Present", resp.Members[1].Nickname)
}

func TestGroupDetailRejectsIncompleteRPCResponse(t *testing.T) {
	logic := NewGroupDetailLogic(context.Background(), &svc.ServiceContext{
		Social: &groupDetailSocialStub{response: &socialclient.GroupDetailResp{}},
	})

	_, err := logic.GroupDetail(&types.GroupDetailReq{GroupId: "1"})
	require.Error(t, err)
}
