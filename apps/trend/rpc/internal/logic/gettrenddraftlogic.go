package logic

import (
	"context"
	"errors"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendDraftLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDraftLogic {
	return &GetTrendDraftLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取动态草稿
func (l *GetTrendDraftLogic) GetTrendDraft(in *trend.GetTrendDraftRequest) (*trend.GetTrendDraftResponse, error) {
	if in.UserId == 0 {
		return nil, errors.New("草稿用户不能为空")
	}

	if in.DraftId > 0 {
		d, err := l.svcCtx.TrendDraft.FindOneByUser(l.ctx, in.UserId, in.DraftId)
		if err != nil {
			return nil, err
		}
		return &trend.GetTrendDraftResponse{Draft: draftToResp(d)}, nil
	}

	d, err := l.svcCtx.TrendDraft.FindLatestByUser(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &trend.GetTrendDraftResponse{Draft: draftToResp(d)}, nil
}
