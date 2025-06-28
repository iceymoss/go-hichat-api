package trend

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteTrendLogic 删除动态
func NewDeleteTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTrendLogic {
	return &DeleteTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTrendLogic) DeleteTrend(req *types.DeleteTrendRequest) (resp *types.DeleteTrendResponse, err error) {
	uid := utils.GetUser(l.ctx)
	res, err := l.svcCtx.Trend.DeleteTrend(l.ctx, &trend.DeleteTrendRequest{
		TrendId: int64(req.TrendID),
		UserId:  int64(uid),
	})
	if err != nil {
		zLog.Error("删除动态失败", zap.Error(err))
		return nil, err
	}

	resp = &types.DeleteTrendResponse{Success: res.Success}

	return
}
