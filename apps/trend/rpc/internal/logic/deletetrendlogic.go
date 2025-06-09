package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
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

// 删除动态
func (l *DeleteTrendLogic) DeleteTrend(in *trend.DeleteTrendRequest) (*trend.DeleteTrendResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.DeleteTrendResponse{}, nil
}
