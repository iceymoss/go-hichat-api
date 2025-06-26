package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTrendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户个人动态
func NewGetUserTrendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTrendsLogic {
	return &GetUserTrendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTrendsLogic) GetUserTrends(req *types.GetUserTrendsRequest) (resp *types.GetUserTrendsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
