package trend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

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

	// 第一次获取需要获取到指定动态
	if req.LastID == 0 {
		topList, topErr := l.svcCtx.Trend.GetUserTopTrend(l.ctx, &trend.GetUserTopTrendsRequest{
			TargetUserId: strconv.Itoa(int(req.TargetUserID)),
			LastId:       0,
		})
		if topErr != nil {
			zLog.Error("获取置顶动态失败", zap.Error(topErr))
			return nil, err
		}
		tops := make([]*types.Trend, 0, len(topList.Trends))
		for _, v := range topList.Trends {
			tops = append(tops, trendRpc2api(v))
		}

		fmt.Println("len:", tops)
		resp.TopTrends = tops
	}

	trend, err := l.svcCtx.Trend.GetUserTrends(l.ctx, &trend.GetUserTrendsRequest{
		TargetUserId: strconv.Itoa(req.TargetUserID),
		Scope:        3, // 全部
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
	list := make([]*types.Trend, 0, len(trend.Trends))
	for _, v := range trend.Trends {
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

	for i, v := range trend.Trends {
		for _, at := range v.AtUserIds {
			list[i].AtUserIds = append(list[i].AtUserIds, userBind[strconv.Itoa(int(at))])
		}
		list[i].User = userBind[strconv.Itoa(req.TargetUserID)]
	}

	resp.Trends = list
	resp.LastID = int(trend.PageInfo.LastId)

	return
}
