package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetTrendDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDetailLogic {
	return &GetTrendDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTrendDetail 获取单个动态详情
func (l *GetTrendDetailLogic) GetTrendDetail(in *trend.GetTrendDetailRequest) (*trend.GetTrendDetailResponse, error) {
	data, err := l.svcCtx.Trend.FindOne(l.ctx, uint64(in.TrendId))
	if err != nil {

		zLog.Error("GetTrendDetail.FindOne: get trend detail failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, err
	}

	return &trend.GetTrendDetailResponse{
		Trend: trendToResp(*data),
	}, nil
}
