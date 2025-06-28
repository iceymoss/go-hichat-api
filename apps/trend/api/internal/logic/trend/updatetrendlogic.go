package trend

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateTrendLogic 更新动态内容
func NewUpdateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTrendLogic {
	return &UpdateTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTrendLogic) UpdateTrend(req *types.UpdateTrendRequest) (resp *types.UpdateTrendResponse, err error) {
	_, err = l.svcCtx.Trend.UpdateTrend(l.ctx, &trend.UpdateTrendRequest{
		TrendId:   req.TrendId,
		IsTop:     req.IsTop,
		OpenReply: req.OpenReply,
		Scope:     req.Scope,
	})
	if err != nil {
		zLog.Error("更新动态失败", zap.Error(err))
		return nil, err
	}

	resp = &types.UpdateTrendResponse{}

	return
}
