package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendLikeSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendLikeSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendLikeSummaryLogic {
	return &GetTrendLikeSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取动态点赞摘要 (用于动态下方展示)
func (l *GetTrendLikeSummaryLogic) GetTrendLikeSummary(in *trend.GetTrendLikeSummaryRequest) (*trend.GetTrendLikeSummaryResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.GetTrendLikeSummaryResponse{}, nil
}
