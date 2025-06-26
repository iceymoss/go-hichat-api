package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除动态
func NewDeleteTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTrendLogic {
	return &DeleteTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTrendLogic) DeleteTrend(req *types.DeleteTrendRequest) (resp *types.DeleteTrendResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
