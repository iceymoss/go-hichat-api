package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetTrendLikeSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetTrendLikeSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetTrendLikeSummaryLogic {
	return &BatchGetTrendLikeSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量获取点赞摘要 (用于动态流)
func (l *BatchGetTrendLikeSummaryLogic) BatchGetTrendLikeSummary(in *trend.BatchTrendLikeSummaryRequest) (*trend.BatchTrendLikeSummaryResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.BatchTrendLikeSummaryResponse{}, nil
}
