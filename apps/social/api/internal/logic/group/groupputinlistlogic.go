package group

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

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
func (l *GroupPutInListLogic) GroupPutInList(req *types.GroupPutInListReq) (resp *types.GroupPutInListResp, err error) {
	uid, err := apiActor(l.ctx)
	if err != nil {
		return nil, err
	}
	res, err := l.svcCtx.Social.GroupPutinList(l.ctx, &social.GroupPutinListReq{
		GroupId: req.GroupId,
		Type:    req.Type,  // []int32
		Class:   req.Class, // int32
		UserId:  uid,
	})
	if err != nil {
		return nil, err
	}

	userList, groupList := make([]string, 0, len(res.List)), make([]string, 0, len(res.List))
	userBindUid, groupBindGid := make(map[string]*user.UserEntity), make(map[string]*social.Groups)
	for _, v := range res.List {
		userList = append(userList, v.ReqId) // ReqId 是发起请求的用户ID
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
		userBindUid[user.Id] = user
	}

	//获取群信息
	groupRes, err := l.svcCtx.Social.FindGroupList(l.ctx, &social.FindGroupListReq{Ids: groupList})
	if err != nil {
		return nil, err
	}

	for _, group := range groupRes.List {
		groupBindGid[group.Id] = group
	}

	list := make([]*types.GroupRequests, 0, len(res.List))
	for _, v := range res.List {
		// 获取请求用户信息（ReqId 是发起请求的用户ID）
		reqUser := userBindUid[v.ReqId]
		if reqUser == nil {
			reqUser = &user.UserEntity{}
		}
		user := types.User{
			Id:           reqUser.Id,
			Nickname:     reqUser.Nickname,
			Sex:          int(reqUser.Sex),
			Avatar:       reqUser.Avatar,
			Introduction: reqUser.Introduction,
		}

		groupInfo := groupBindGid[v.GroupId]
		if groupInfo == nil {
			groupInfo = &social.Groups{}
		}
		group := types.Groups{
			Id:        groupInfo.Id,
			Name:      groupInfo.Name,
			Icon:      groupInfo.Icon,
			Status:    int64(groupInfo.Status),
			CreateUid: groupInfo.CreatorUid,
		}
		list = append(list, &types.GroupRequests{
			Id:                 int64(v.Id),
			UserId:             v.ReqId, // 请求用户ID
			GroupId:            v.GroupId,
			User:               user,
			Group:              group,
			ReqMsg:             v.ReqMsg,
			ReqTime:            v.ReqTime,
			JoinSource:         int64(v.JoinSource),
			InviterUserId:      v.InviterUid,
			HandleUserId:       v.HandleUid,
			HandleTime:         v.HandleResultTime,
			HandleResult:       int64(v.HandleResult),
			ReceiverRead:       int64(v.ReceiverRead),
			RequestId:          strconv.FormatUint(v.RequestId, 10),
			ApplicantUid:       v.ApplicantUid,
			HandleMsg:          v.HandleMsg,
			InvalidReason:      v.InvalidReason,
			ActualJoinSource:   v.ActualJoinSource,
			SourceType:         v.SourceType,
			SourceInvitationId: formatOptionalID(v.SourceInvitationId),
			ReadState:          v.ReadState,
			Actionable:         v.Actionable,
		})
	}

	resp = &types.GroupPutInListResp{
		List:  list,
		Total: res.Total,
	}

	return
}
