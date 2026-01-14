package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupDetailLogic {
	return &GroupDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupDetailLogic) GroupDetail(req *types.GroupDetailReq) (resp *types.GroupDetailResp, err error) {
	rpcResp, err := l.svcCtx.Social.GroupDetail(l.ctx, &social.GroupDetailReq{
		GroupId: req.GroupId,
	})
	if err != nil {
		return nil, err
	}

	// enrich member user profile via user-rpc
	userIdList := make([]string, 0, len(rpcResp.Members))
	for _, m := range rpcResp.Members {
		userIdList = append(userIdList, m.UserId)
	}

	userBindUid := make(map[string]user.UserEntity, len(userIdList))
	if len(userIdList) > 0 {
		userRes, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{Ids: userIdList})
		if err != nil {
			return nil, err
		}
		for _, u := range userRes.User {
			userBindUid[u.Id] = *u
		}
	}

	members := make([]*types.GroupMembers, 0)
	for _, m := range rpcResp.Members {
		u := userBindUid[m.UserId]
		members = append(members, &types.GroupMembers{
			Id:          int64(m.Id),
			GroupId:     m.GroupId,
			UserId:      m.UserId,
			Nickname:    u.Nickname,
			UserAvatarUrl: u.Avatar,
			User: types.User{
				Id:           u.Id,
				Nickname:     u.Nickname,
				Sex:          int(u.Sex),
				Avatar:       u.Avatar,
				Introduction: u.Introduction,
			},
			RoleLevel:   int(m.RoleLevel),
			InviterUid:  m.InviterUid,
			OperatorUid: m.OperatorUid,
		})
	}

	return &types.GroupDetailResp{
		Group: types.Groups{
			Id:              rpcResp.Group.Id,
			Name:            rpcResp.Group.Name,
			Icon:            rpcResp.Group.Icon,
			Status:          int64(rpcResp.Group.Status),
			CreateUid:       rpcResp.Group.CreatorUid,
			GroupType:       int64(rpcResp.Group.GroupType),
			IsVerify:        rpcResp.Group.IsVerify,
			Notification:    rpcResp.Group.Notification,
			NotificationUid: rpcResp.Group.NotificationUid,
		},
		Members: members,
	}, nil
}
