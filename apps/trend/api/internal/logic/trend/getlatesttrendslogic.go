package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTrendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取最新动态流
func NewGetLatestTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTrendsLogic {
	return &GetLatestTrendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLatestTrendsLogic) GetLatestTrends(req *types.GetLatestTrendsRequest) (resp *types.GetLatestTrendsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
