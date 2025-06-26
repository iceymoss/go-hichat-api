package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新动态内容
func NewUpdateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTrendLogic {
	return &UpdateTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTrendLogic) UpdateTrend(req *types.UpdateTrendRequest) (resp *types.UpdateTrendResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
