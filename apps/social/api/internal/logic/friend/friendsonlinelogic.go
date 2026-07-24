package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendsOnlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendsOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendsOnlineLogic {
	return &FriendsOnlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendsOnlineLogic) FriendsOnline(req *types.FriendsOnlineReq) (resp *types.FriendsOnlineResp, err error) {
	uid := ctxdata.GetUId(l.ctx)

	// 1. 获取好友列表
	friendListResp, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{
		UserId: uid,
	})
	if err != nil {
		return nil, err
	}

	onlineMap := make(map[string]bool)

	if len(friendListResp.List) == 0 {
		return &types.FriendsOnlineResp{OnlineList: onlineMap}, nil
	}

	// 2. 批量查 Redis 在线状态
	friendIds := make([]string, 0, len(friendListResp.List))

	for _, f := range friendListResp.List {
		friendIds = append(friendIds, f.FriendUid)
	}

	onlineMap, err = l.svcCtx.Presence.BatchOnline(l.ctx, friendIds)
	if err != nil {
		l.Errorf("query friend presence: %v", err)
		onlineMap = make(map[string]bool, len(friendIds))
		for _, id := range friendIds {
			onlineMap[id] = false
		}
		return &types.FriendsOnlineResp{OnlineList: onlineMap}, nil
	}

	return &types.FriendsOnlineResp{OnlineList: onlineMap}, nil
}
