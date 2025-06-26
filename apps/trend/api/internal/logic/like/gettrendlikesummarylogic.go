package like

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendLikeSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取动态点赞摘要
func NewGetTrendLikeSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendLikeSummaryLogic {
	return &GetTrendLikeSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTrendLikeSummaryLogic) GetTrendLikeSummary(req *types.GetTrendLikeSummaryRequest) (resp *types.GetTrendLikeSummaryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
