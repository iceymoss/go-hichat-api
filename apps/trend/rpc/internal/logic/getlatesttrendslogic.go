package logic

import (
	"context"
	"errors"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTrendsLogic {
	return &GetLatestTrendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetLatestTrends 获取最新动态
func (l *GetLatestTrendsLogic) GetLatestTrends(in *trend.GetLatestTrendsRequest) (*trend.GetLatestTrendsResponse, error) {
	list, err := l.svcCtx.Trend.List(l.ctx, int(in.LastTrendId), 0, in.UserIds, []string{"*"}, "id", -1)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetLatestTrends.List: %v", err))
	}

	if len(*list) > 0 {

	}

	return &trend.GetLatestTrendsResponse{}, nil
}
