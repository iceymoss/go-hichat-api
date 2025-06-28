package trend

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTrendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLatestTrendsLogic 获取最新动态流
func NewGetLatestTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTrendsLogic {
	return &GetLatestTrendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLatestTrendsLogic) GetLatestTrends(req *types.GetLatestTrendsRequest) (resp *types.GetLatestTrendsResponse, err error) {
	uid := utils.GetUser(l.ctx)

	// 获取用户好友
	friendList, err := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{UserId: strconv.Itoa(uid)})
	if err != nil {
		zLog.Error("获取好友列表失败", zap.Error(err))
		return nil, err
	}
	friendIdList := make([]string, 0, len(friendList.List)+1)
	friendIdList = append(friendIdList, strconv.Itoa(uid))
	for _, v := range friendList.List {
		friendIdList = append(friendIdList, v.FriendUid)
	}

	trendLast, err := l.svcCtx.Trend.GetLatestTrends(l.ctx, &trend.GetLatestTrendsRequest{
		UserIds:     friendIdList,
		LastTrendId: int64(req.LastTrendID),
		Count:       int32(req.Count),
	})
	if err != nil {
		zLog.Error("获取最新动态失败", zap.Error(err))
		return nil, err
	}

	// 获取被@的用户信息
	atUidInt := make([]int32, 0)
	for _, v := range trendLast.Trends {
		atUidInt = append(atUidInt, v.AtUserIds...)
	}

	atUid := make([]string, 0, len(atUidInt))
	for _, userId := range atUidInt {
		atUid = append(atUid, strconv.Itoa(int(userId)))
	}
	friendIdList = append(friendIdList, atUid...)

	// 获取用户信息
	userInfoBind := make(map[string]*types.User, len(friendIdList))
	userInfoResp, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: friendIdList,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Error(err))
		return nil, err
	}

	for _, v := range userInfoResp.User {
		userInfoBind[v.Id] = &types.User{
			Id:       v.Id,
			Nickname: v.Nickname,
			Sex:      int(v.Sex),
			Avatar:   v.Avatar,
		}
	}

	list := make([]*types.Trend, 0, len(trendLast.Trends))
	for _, v := range trendLast.Trends {
		trendTemp := trendRpc2api(v)
		trendTemp.User = userInfoBind[strconv.Itoa(int(v.UserId))]
		atUserInfo := make([]*types.User, 0, len(v.AtUserIds))
		for _, userInfo := range v.AtUserIds {
			if user, ok := userInfoBind[strconv.Itoa(int(userInfo))]; ok {
				atUserInfo = append(atUserInfo, user)
			}
		}
		trendTemp.AtUserIds = atUserInfo
		list = append(list, trendTemp)
	}

	resp = &types.GetLatestTrendsResponse{
		Trends:      list,
		LastTrendID: int(trendLast.LastTrendId),
		HasMore:     trendLast.HasMore,
	}

	return
}

func trendRpc2api(trend *trend.Trend) *types.Trend {
	return &types.Trend{
		Id:            trend.Id,
		Type:          int(trend.Type),
		Content:       trend.Content,
		Scope:         int(trend.Scope),
		CreateTime:    trend.CreateTime,
		ReplyCount:    trend.ReplyCount,
		AgreeCount:    trend.AgreeCount,
		PositionName:  trend.PositionName,
		PositionPoint: trend.PositionPoint,
		Title:         trend.Title,
		Resources:     trend.Resources,
		CoverUrl:      trend.CoverUrl,
		ShareUrl:      trend.ShareUrl,
		OpenReply:     trend.OpenReply,
		DeviceId:      trend.DeviceId,
		Ip:            trend.Ip,
		IsTop:         int(trend.IsTop),
	}
}
