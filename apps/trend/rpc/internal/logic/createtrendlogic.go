package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTrendLogic {
	return &CreateTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布动态
func (l *CreateTrendLogic) CreateTrend(in *trend.CreateTrendRequest) (*trend.CreateTrendResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.CreateTrendResponse{}, nil
}
