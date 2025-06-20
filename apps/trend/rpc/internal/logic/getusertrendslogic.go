package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetUserTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTrendsLogic {
	return &GetUserTrendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserTrends 获取用户个人动态列表
func (l *GetUserTrendsLogic) GetUserTrends(in *trend.GetUserTrendsRequest) (*trend.GetUserTrendsResponse, error) {
	list, err := l.svcCtx.Trend.ListByUserIds(l.ctx, []string{in.TargetUserId}, int(in.Pagination.LastId), int32(in.Scope), []string{"*"})
	if err != nil {
		zLog.Error("GetUserTrends.ListByUserIds", zap.Any("userId", in.TargetUserId), zap.Any("lastTime", in.Pagination.LastTime), zap.Any("scope", in.Scope), zap.Error(err))
		return nil, err
	}

	trendList := make([]*trend.Trend, 0, len(*list))
	for _, v := range *list {
		trendList = append(trendList, trendToResp(v))
	}

	last := &trend.Trend{}
	if len(trendList) > 0 {
		last = trendList[len(trendList)-1]
	}

	// TODO: 获取点赞和评论数据

	return &trend.GetUserTrendsResponse{
		Trends: trendList,
		PageInfo: &trend.PageInfo{
			LastId:   int32(last.Id),
			LastTime: last.CreateTime,
		},
	}, nil
}
