package friend

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 好友列表
func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendListLogic) FriendList(req *types.FriendListReq) (resp *types.FriendListResp, err error) {
	uid := l.ctx.Value(Identify).(string)
	res, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{
		UserId: uid,
	})
	if err != nil || res == nil {
		return nil, err
	}

	ids := make([]string, 0, len(res.List))
	for _, v := range res.List {
		ids = append(ids, v.FriendUid)
	}
	//获取用户信息
	userList, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Name:  "",
		Phone: "",
		Ids:   ids,
	})
	if err != nil || userList == nil {
		return nil, err
	}

	uidBindInfo := make(map[string]*user.UserEntity)
	for _, v := range userList.User {
		uidBindInfo[v.Id] = v
	}

	list := make([]*types.Friends, 0, len(userList.User))
	for _, v := range res.List {
		if user, ok := uidBindInfo[v.FriendUid]; ok {
			if v.Remark == "" {
				v.Remark = user.Nickname
			}
			item := &types.Friends{
				Id:        v.Id,
				FriendUid: v.FriendUid,
				Nickname:  user.Nickname,
				Avatar:    user.Avatar,
				Remark:    v.Remark,
			}
			list = append(list, item)
		}
	}

	resp = &types.FriendListResp{List: list}

	return
}
