package logic

import (
	"context"
	"errors"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTrendDraftLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTrendDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTrendDraftLogic {
	return &DeleteTrendDraftLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除动态草稿
func (l *DeleteTrendDraftLogic) DeleteTrendDraft(in *trend.DeleteTrendDraftRequest) (*trend.DeleteTrendDraftResponse, error) {
	if in.UserId == 0 || in.DraftId == 0 {
		return nil, errors.New("草稿参数不能为空")
	}
	if err := l.svcCtx.TrendDraft.SoftDelete(l.ctx, in.UserId, in.DraftId); err != nil {
		return nil, err
	}

	return &trend.DeleteTrendDraftResponse{Success: true}, nil
}
