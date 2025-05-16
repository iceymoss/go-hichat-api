package group

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGroupPutInListLogic 申请进群列表
func NewGroupPutInListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInListLogic {
	return &GroupPutInListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GroupPutInList 获取群申请，或者获取用户发起的申请
// 获取用户发起的申请；
//
//	{
//		"group_id": "0",
//		"type": [2],
//		"class": 1
//	}
//
// 获取某一个群的加群申请：
//
//	{
//		"group_id": "13",
//		"type": [],
//		"class": 2
//	}
func (l *GroupPutInListLogic) GroupPutInList(req *types.GroupPutInListRep) (resp *types.GroupPutInListResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.Social.GroupPutinList(l.ctx, &social.GroupPutinListReq{
		GroupId: req.GroupId,
		Type:    req.Type,
		Class:   int32(req.Class),
		UserId:  uid,
	})
	if err != nil {
		return nil, err
	}

	userList, groupList := make([]string, 0, len(res.List)), make([]string, 0, len(res.List))
	userBindUid, groupBindGid := make(map[string]user.UserEntity), make(map[string]social.Groups)
	for _, v := range res.List {
		userList = append(userList, v.ReqId)
		groupList = append(groupList, v.GroupId)
	}

	//获取用户信息
	userRes, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: userList,
	})
	if err != nil {
		return nil, err
	}

	for _, user := range userRes.User {
		userBindUid[user.Id] = *user
	}

	//获取群信息
	groupRes, err := l.svcCtx.Social.FindGroupList(l.ctx, &social.FindGroupListReq{Ids: groupList})
	if err != nil {
		return nil, err
	}

	for _, group := range groupRes.List {
		groupBindGid[group.Id] = *group
	}

	list := make([]*types.GroupRequests, 0, len(res.List))
	for _, v := range res.List {
		user := types.User{
			Id:           userBindUid[v.HandleUid].Id,
			Nickname:     userBindUid[v.HandleUid].Nickname,
			Sex:          int(userBindUid[v.HandleUid].Sex),
			Avatar:       userBindUid[v.HandleUid].Avatar,
			Introduction: userBindUid[v.HandleUid].Introduction,
		}

		group := types.Groups{
			Id:        groupBindGid[v.GroupId].Id,
			Name:      groupBindGid[v.GroupId].Name,
			Icon:      groupBindGid[v.GroupId].Icon,
			Status:    int64(groupBindGid[v.GroupId].Status),
			CreateUid: groupBindGid[v.GroupId].CreatorUid,
		}
		list = append(list, &types.GroupRequests{
			Id:            int64(v.Id),
			User:          user,
			Group:         group,
			ReqMsg:        v.ReqMsg,
			ReqTime:       v.ReqTime,
			JoinSource:    int64(v.JoinSource),
			InviterUserId: v.InviterUid,
			HandleUserId:  v.HandleUid,
			HandleTime:    v.HandleResultTime,
			HandleResult:  int64(v.HandleResult),
		})
	}

	resp = &types.GroupPutInListResp{
		List: list,
	}

	return
}
