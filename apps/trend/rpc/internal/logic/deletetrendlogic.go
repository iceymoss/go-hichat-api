package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type DeleteTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTrendLogic {
	return &DeleteTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteTrend 删除动态
func (l *DeleteTrendLogic) DeleteTrend(in *trend.DeleteTrendRequest) (*trend.DeleteTrendResponse, error) {
	//trendData, err := l.svcCtx.Trend.FindOne(l.ctx, uint64(in.TrendId))
	//if err != nil {
	//	zLog.Error("DeleteTrend.FindOne: get trend detail failed", zap.Any("trendId", in.TrendId), zap.Error(err))
	//	return nil, err
	//}
	//if trendData.Userid != uint64(in.UserId) {
	//	return nil, errors.New("do not have permission")
	//}
	err := l.svcCtx.Trend.Delete(l.ctx, uint64(in.TrendId))
	if err != nil {
		zLog.Error("DeleteTrend.Delete: delete trend detail failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, err
	}

	return &trend.DeleteTrendResponse{
		Success: true,
	}, nil
}
