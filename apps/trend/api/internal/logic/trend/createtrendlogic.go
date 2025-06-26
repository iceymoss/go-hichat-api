package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建新动态
func NewCreateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTrendLogic {
	return &CreateTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTrendLogic) CreateTrend(req *types.CreateTrendRequest) (resp *types.CreateTrendResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
