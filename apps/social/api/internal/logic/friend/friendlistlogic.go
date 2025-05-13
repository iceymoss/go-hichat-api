package friend

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"

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
	// todo: add your logic here and delete this line

	//todo: 从token获取用户uid

	uid := "11"
	res, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{
		UserId: uid,
	})
	if err != nil {
		return
	}

	list := make([]*types.Friends, 0, len(res.List))
	for _, v := range res.List {
		list = append(list, &types.Friends{
			Id:        v.Id,
			FriendUid: v.FriendUid,
			Nickname:  "",
			Avatar:    "",
			Remark:    v.Remark,
		})
	}

	resp.List = list

	return
}
