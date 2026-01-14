package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupAnnouncementListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupAnnouncementListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAnnouncementListLogic {
	return &GroupAnnouncementListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupAnnouncementListLogic) GroupAnnouncementList(req *types.GroupAnnouncementListReq) (resp *types.GroupAnnouncementListResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GroupAnnouncementList(l.ctx, &social.GroupAnnouncementListReq{
		UserId:         uid,
		GroupId:        req.GroupId,
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rpcResp.List))
	for _, a := range rpcResp.List {
		ids = append(ids, a.CreatedBy)
	}

	userBind := map[string]user.UserEntity{}
	if len(ids) > 0 {
		userRes, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{Ids: ids})
		if err != nil {
			return nil, err
		}
		for _, u := range userRes.User {
			userBind[u.Id] = *u
		}
	}

	list := make([]*types.GroupAnnouncement, 0, len(rpcResp.List))
	for _, a := range rpcResp.List {
		u := userBind[a.CreatedBy]
		var creator *types.User
		if u.Id != "" {
			creator = &types.User{
				Id:           u.Id,
				Nickname:     u.Nickname,
				Sex:          int(u.Sex),
				Avatar:       u.Avatar,
				Introduction: u.Introduction,
			}
		}

		list = append(list, &types.GroupAnnouncement{
			Id:        a.Id,
			GroupId:   a.GroupId,
			Content:   a.Content,
			CreatedBy: a.CreatedBy,
			CreatedAt: a.CreatedAt,
			Pinned:    a.Pinned,
			PinnedAt:  a.PinnedAt,
			Creator:   creator,
		})
	}

	return &types.GroupAnnouncementListResp{List: list}, nil
}
