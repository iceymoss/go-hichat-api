package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	trendrpc "github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTrendDraftLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除动态草稿
func NewDeleteTrendDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTrendDraftLogic {
	return &DeleteTrendDraftLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTrendDraftLogic) DeleteTrendDraft(req *types.DeleteTrendDraftRequest) (resp *types.DeleteTrendDraftResponse, err error) {
	uid := utils.GetUser(l.ctx)
	res, err := l.svcCtx.Trend.DeleteTrendDraft(l.ctx, &trendrpc.DeleteTrendDraftRequest{
		UserId:  uint64(uid),
		DraftId: req.DraftId,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteTrendDraftResponse{Success: res.Success}, nil
}
