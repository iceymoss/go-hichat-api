package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTrendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrendsLogic {
	return &ListTrendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取动态列表
func (l *ListTrendsLogic) ListTrends(in *trend.ListTrendsRequest) (*trend.ListTrendsResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.ListTrendsResponse{}, nil
}
