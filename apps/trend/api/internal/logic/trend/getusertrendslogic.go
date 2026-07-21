package trend

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUserTrendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUserTrendsLogic 获取用户个人动态
func NewGetUserTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTrendsLogic {
	return &GetUserTrendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTrendsLogic) GetUserTrends(req *types.GetUserTrendsRequest) (resp *types.GetUserTrendsResponse, err error) {
	resp = &types.GetUserTrendsResponse{
		Trends:    nil,
		LastID:    0,
		TopTrends: nil,
	}

	currentUID := utils.GetUser(l.ctx)
	isSelf := req.TargetUserID == currentUID

	// 看他人朋友圈: 必须是好友, 否则拒绝访问
	if !isSelf {
		friendList, ferr := l.svcCtx.Social.FriendList(l.ctx, &social.FriendListReq{UserId: strconv.Itoa(currentUID)})
		if ferr != nil {
			zLog.Error("获取好友列表失败", zap.Error(ferr))
			return nil, ferr
		}
		target := strconv.Itoa(req.TargetUserID)
		isFriend := false
		for _, v := range friendList.List {
			if v.FriendUid == target {
				isFriend = true
				break
			}
		}
		if !isFriend {
			return nil, errors.New("对方不是您的好友, 无法查看其朋友圈")
		}
	}

	// scope 透传到 model.ListByUserIds 的 circle_state 过滤:
	// 3 -> circle_state in (1,2) 全部(含仅自己可见); 1 -> circle_state=1 仅对外可见
	// 看自己返回全部, 看好友只返回对外可见的动态
	listScope := trend.VisibilityScope(3)
	if !isSelf {
		listScope = trend.VisibilityScope(1)
	}

	// 第一次获取需要获取到置顶动态
	if req.LastID == 0 {
		topList, topErr := l.svcCtx.Trend.GetUserTopTrend(l.ctx, &trend.GetUserTopTrendsRequest{
			TargetUserId: strconv.Itoa(req.TargetUserID),
			LastId:       0,
		})
		if topErr != nil {
			zLog.Error("获取置顶动态失败", zap.Error(topErr))
			return nil, topErr
		}
		tops := make([]*types.Trend, 0, len(topList.Trends))
		for _, v := range topList.Trends {
			// 置顶查询不带 circle_state 过滤, 看他人时需手动剔除对方"仅自己可见"的置顶
			if !isSelf && v.Scope == trend.VisibilityScope_PRIVATE {
				continue
			}
			tops = append(tops, trendRpc2api(v))
		}

		resp.TopTrends = tops
	}

	userTrends, err := l.svcCtx.Trend.GetUserTrends(l.ctx, &trend.GetUserTrendsRequest{
		TargetUserId: strconv.Itoa(req.TargetUserID),
		Scope:        listScope,
		Pagination: &trend.Pagination{
			LastId: int32(req.LastID),
		},
	})
	if err != nil {
		zLog.Error("获取用户动态失败", zap.Error(err))
		return nil, err
	}

	// 用户信息
	uidList := []string{strconv.Itoa(req.TargetUserID)}
	list := make([]*types.Trend, 0, len(userTrends.Trends))
	for _, v := range userTrends.Trends {
		trendTemp := trendRpc2api(v)
		list = append(list, trendTemp)
		for _, atUid := range v.AtUserIds {
			uidList = append(uidList, strconv.Itoa(int(atUid)))
		}
	}

	userInfo, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: uidList,
	})
	if err != nil {
		zLog.Error("获取用户信息失败", zap.Error(err))
		return nil, err
	}

	userBind := make(map[string]*types.User, len(userInfo.User))
	for _, v := range userInfo.User {
		userBind[v.Id] = &types.User{
			Id:       v.Id,
			Nickname: v.Nickname,
			Sex:      int(v.Sex),
			Avatar:   v.Avatar,
		}
	}

	for i, v := range userTrends.Trends {
		for _, at := range v.AtUserIds {
			list[i].AtUserIds = append(list[i].AtUserIds, userBind[strconv.Itoa(int(at))])
		}
		list[i].User = userBind[strconv.Itoa(req.TargetUserID)]
	}

	resp.Trends = list
	resp.LastID = int(userTrends.PageInfo.LastId)
	resp.LastTime = int(userTrends.PageInfo.LastTime)

	return
}
