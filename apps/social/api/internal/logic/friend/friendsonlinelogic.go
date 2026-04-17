package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

const onlineKeyPrefix = "user:online:"

// Package-level Redis client (lazy init, reuse connection)
var rdb *redis.Client

func getRedis() *redis.Client {
	if rdb == nil {
		rdb = redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:6379",
		})
	}
	return rdb
}

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
	r := getRedis()
	keys := make([]string, 0, len(friendListResp.List))
	friendIds := make([]string, 0, len(friendListResp.List))

	for _, f := range friendListResp.List {
		friendIds = append(friendIds, f.FriendUid)
		keys = append(keys, onlineKeyPrefix+f.FriendUid)
	}

	results, err := r.MGet(l.ctx, keys...).Result()
	if err != nil {
		// Redis 错误不影响接口，返回全部离线
		return &types.FriendsOnlineResp{OnlineList: onlineMap}, nil
	}

	for i, val := range results {
		onlineMap[friendIds[i]] = val != nil
	}

	return &types.FriendsOnlineResp{OnlineList: onlineMap}, nil
}
