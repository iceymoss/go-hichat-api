package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTrendLogic {
	return &UpdateTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新动态
func (l *UpdateTrendLogic) UpdateTrend(in *trend.UpdateTrendRequest) (*trend.UpdateTrendResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.UpdateTrendResponse{}, nil
}
