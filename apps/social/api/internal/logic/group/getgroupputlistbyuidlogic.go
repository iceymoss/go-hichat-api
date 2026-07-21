package group

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupPutListByUidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetGroupPutListByUidLogic 用户角度的群申请列表
func NewGetGroupPutListByUidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupPutListByUidLogic {
	return &GetGroupPutListByUidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetGroupPutListByUid 用户角度的群申请列表
// ids: 用户id列表（可选，如果不传则使用当前登录用户）
// class: 类别：1-我发起的申请，2-我接受到的申请
// type: 状态：0-未处理，1-已通过，2-已拒绝，3-已忽略
func (l *GetGroupPutListByUidLogic) GetGroupPutListByUid(req *types.GetGroupPutListByUidReq) (resp *types.GetGroupPutListByUidResp, err error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}
	// Public callers are always scoped to the JWT actor; legacy ids cannot expand visibility.
	ids := []string{uid}

	// 调用RPC
	res, err := l.svcCtx.Social.GetGroupPutListByUid(l.ctx, &social.GetGroupPutListByUidReq{
		Ids:      ids,
		ActorUid: uid,
		Class:    req.Class,
		Type:     req.Type,
		Status:   req.Status,
		Page:     req.Page,
		Size:     req.Size,
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

	// 获取用户信息
	if len(userList) > 0 {
		userRes, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
			Ids: userList,
		})
		if err != nil {
			return nil, err
		}

		for _, user := range userRes.User {
			userBindUid[user.Id] = user
		}
	}

	// 获取群信息
	if len(groupList) > 0 {
		groupRes, err := l.svcCtx.Social.FindGroupList(l.ctx, &social.FindGroupListReq{Ids: groupList})
		if err != nil {
			return nil, err
		}

		for _, group := range groupRes.List {
			groupBindGid[group.Id] = group
		}
	}

	list := make([]*types.GroupRequests, 0, len(res.List))
	for _, v := range res.List {
		// 获取请求用户信息（ReqId 是发起请求的用户ID）
		reqUser := userBindUid[v.ReqId]
		if reqUser == nil {
			reqUser = &user.UserEntity{}
		}
		groupInfo := groupBindGid[v.GroupId]
		if groupInfo == nil {
			groupInfo = &social.Groups{}
		}
		user := types.User{
			Id:           reqUser.Id,
			Nickname:     reqUser.Nickname,
			Sex:          int(reqUser.Sex),
			Avatar:       reqUser.Avatar,
			Introduction: reqUser.Introduction,
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

	resp = &types.GetGroupPutListByUidResp{
		List: list, Total: res.Total,
	}

	return
}

func formatOptionalID(id uint64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(id, 10)
}
