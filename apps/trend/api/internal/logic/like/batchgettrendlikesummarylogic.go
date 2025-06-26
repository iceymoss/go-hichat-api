package like

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetTrendLikeSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量获取点赞摘要
func NewBatchGetTrendLikeSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetTrendLikeSummaryLogic {
	return &BatchGetTrendLikeSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchGetTrendLikeSummaryLogic) BatchGetTrendLikeSummary(req *types.BatchTrendLikeSummaryRequest) (resp *types.BatchTrendLikeSummaryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
