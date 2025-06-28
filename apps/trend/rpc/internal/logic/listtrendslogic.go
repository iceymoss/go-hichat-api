package logic

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type ListTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrendsLogic {
	return &ListTrendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListTrends 获取动态列表
func (l *ListTrendsLogic) ListTrends(in *trend.ListTrendsRequest) (*trend.ListTrendsResponse, error) {
	list, err := l.svcCtx.Trend.List(l.ctx, int(in.Pagination.LastId), 0, in.UserIds, []string{"*"}, "id", -1)
	if err != nil {
		zLog.Error("ListTrends.List: get trend list filed", zap.Any("lastId", "in.Pagination.LastId"), zap.Any("lastTime", in.Pagination.LastTime))
		return nil, err
	}

	trendList := make([]*trend.Trend, 0, len(*list))
	for _, v := range *list {
		trendList = append(trendList, trendToResp(v))
	}

	last := &trend.Trend{}
	if len(trendList) > 0 {
		last = trendList[len(trendList)-1]
		fmt.Printf("data: %+v\n", last)
	}

	return &trend.ListTrendsResponse{
		Trends: trendList,
		PageInfo: &trend.PageInfo{
			LastId:   int32(last.Id),
			LastTime: last.CreateTime,
		},
	}, nil
}
