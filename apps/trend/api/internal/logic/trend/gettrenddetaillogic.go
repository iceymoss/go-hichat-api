package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取动态详情
func NewGetTrendDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDetailLogic {
	return &GetTrendDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTrendDetailLogic) GetTrendDetail(req *types.GetTrendDetailRequest) (resp *types.GetTrendDetailResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
