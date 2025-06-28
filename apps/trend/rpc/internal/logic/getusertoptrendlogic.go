package logic

import (
	"context"
	"gorm.io/gorm"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUserTopTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTopTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTopTrendLogic {
	return &GetUserTopTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserTopTrend 获取用户动态置顶
func (l *GetUserTopTrendLogic) GetUserTopTrend(in *trend.GetUserTopTrendsRequest) (*trend.GetUserTrendsResponse, error) {
	uid, err := strconv.Atoi(in.TargetUserId)
	if err != nil {
		zLog.Error("用户id非法", zap.Error(err))
		return nil, err
	}
	trendList, err := l.svcCtx.Trend.GetTopByUid(l.ctx, uint64(uid), uint64(in.LastId))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &trend.GetUserTrendsResponse{
				Trends: []*trend.Trend{},
				PageInfo: &trend.PageInfo{
					LastId: 0,
				},
			}, nil
		}
		zLog.Error("获取用户动态置顶失败", zap.Any("uid", uid), zap.Any("targetUserId", in.TargetUserId), zap.Error(err))
		return nil, err
	}

	list := make([]*trend.Trend, 0, len(trendList))
	for _, v := range trendList {
		list = append(list, trendToResp(*v))
	}

	var lastId int32
	if len(list) > 0 {
		lastId = int32(list[len(list)-1].Id)
	}

	return &trend.GetUserTrendsResponse{
		Trends: list,
		PageInfo: &trend.PageInfo{
			LastId: lastId,
		},
	}, nil
}
