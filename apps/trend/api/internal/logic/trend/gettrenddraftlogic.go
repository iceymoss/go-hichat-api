package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	trendrpc "github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendDraftLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取动态草稿
func NewGetTrendDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDraftLogic {
	return &GetTrendDraftLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTrendDraftLogic) GetTrendDraft(req *types.GetTrendDraftRequest) (resp *types.GetTrendDraftResponse, err error) {
	uid := utils.GetUser(l.ctx)
	res, err := l.svcCtx.Trend.GetTrendDraft(l.ctx, &trendrpc.GetTrendDraftRequest{
		UserId:  uint64(uid),
		DraftId: req.DraftId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetTrendDraftResponse{Draft: rpcDraftToAPI(res.Draft)}, nil
}
