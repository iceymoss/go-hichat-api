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

type GroupAtListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupAtListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAtListLogic {
	return &GroupAtListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupAtListLogic) GroupAtList(req *types.GroupAtListReq) (resp *types.GroupAtListResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GroupAtList(l.ctx, &social.GroupAtListReq{
		UserId:  uid,
		GroupId: req.GroupId,
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rpcResp.List))
	for _, m := range rpcResp.List {
		ids = append(ids, m.UserId)
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

	list := make([]*types.AtMember, 0, len(rpcResp.List))
	for _, m := range rpcResp.List {
		u := userBind[m.UserId]
		list = append(list, &types.AtMember{
			UserId:        m.UserId,
			Nickname:      u.Nickname,
			GroupNickname: m.GroupNickname,
			Avatar:        u.Avatar,
			RoleLevel:     m.RoleLevel,
		})
	}

	return &types.GroupAtListResp{List: list}, nil
}
