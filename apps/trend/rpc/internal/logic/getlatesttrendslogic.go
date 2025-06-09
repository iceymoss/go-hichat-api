package logic

import (
	"context"

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

// 获取最新动态（用于朋友圈/论坛动态流）
func (l *GetLatestTrendsLogic) GetLatestTrends(in *trend.GetLatestTrendsRequest) (*trend.GetLatestTrendsResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.GetLatestTrendsResponse{}, nil
}
