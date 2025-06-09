package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDetailLogic {
	return &GetTrendDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个动态详情
func (l *GetTrendDetailLogic) GetTrendDetail(in *trend.GetTrendDetailRequest) (*trend.GetTrendDetailResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.GetTrendDetailResponse{}, nil
}
