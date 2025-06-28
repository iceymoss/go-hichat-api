package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTrendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListTrendsLogic 获取动态列表
func NewListTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrendsLogic {
	return &ListTrendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTrendsLogic) ListTrends(req *types.ListTrendsRequest) (resp *types.ListTrendsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
