package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTrendsLogic {
	return &GetUserTrendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户个人动态列表
func (l *GetUserTrendsLogic) GetUserTrends(in *trend.GetUserTrendsRequest) (*trend.GetUserTrendsResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.GetUserTrendsResponse{}, nil
}
