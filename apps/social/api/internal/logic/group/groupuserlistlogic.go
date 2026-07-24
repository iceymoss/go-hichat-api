package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 成员列表列表
func NewGroupUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUserListLogic {
	return &GroupUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupUserListLogic) GroupUserList(req *types.GroupUserListReq) (resp *types.GroupUserListResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	groupMember, err := l.svcCtx.Social.GroupUsers(l.ctx, &social.GroupUsersReq{GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}

	//get user info
	userIdList := make([]string, 0, len(groupMember.List))
	for _, m := range groupMember.List {
		if m == nil {
			continue
		}
		userIdList = append(userIdList, m.UserId)
	}

	//获取用户信息
	userBindUid := make(map[string]*user.UserEntity)
	userRes, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: userIdList,
	})
	if err != nil {
		return nil, err
	}

	for _, user := range userRes.User {
		if user == nil {
			continue
		}
		userBindUid[user.Id] = user
	}

	list := make([]*types.GroupMembers, 0, len(userIdList))
	for _, m := range groupMember.List {
		if m == nil {
			continue
		}
		var IsCurrentUser int
		if m.UserId == uid {
			IsCurrentUser = 1
		}
		userInfo := userBindUid[m.UserId]
		if userInfo == nil {
			userInfo = &user.UserEntity{}
		}
		list = append(list, &types.GroupMembers{
			Id:            int64(m.Id),
			GroupId:       m.GroupId,
			UserId:        m.UserId,
			Nickname:      userInfo.Nickname,
			UserAvatarUrl: userInfo.Avatar,
			User: types.User{
				Id:            userInfo.Id,
				Nickname:      userInfo.Nickname,
				Sex:           int(userInfo.Sex),
				Avatar:        userInfo.Avatar,
				Introduction:  userInfo.Introduction,
				IsCurrentUser: IsCurrentUser,
			},
			RoleLevel:  int(m.RoleLevel),
			InviterUid: m.InviterUid,
		})
	}

	resp = &types.GroupUserListResp{
		List: list,
	}

	return
}
