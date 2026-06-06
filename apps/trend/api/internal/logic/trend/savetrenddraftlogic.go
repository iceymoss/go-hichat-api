package trend

import (
	"context"
	"errors"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	trendrpc "github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveTrendDraftLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存动态草稿
func NewSaveTrendDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveTrendDraftLogic {
	return &SaveTrendDraftLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveTrendDraftLogic) SaveTrendDraft(req *types.SaveTrendDraftRequest) (resp *types.SaveTrendDraftResponse, err error) {
	if req == nil || req.Draft == nil {
		return nil, errors.New("草稿不能为空")
	}
	uid := utils.GetUser(l.ctx)
	res, err := l.svcCtx.Trend.SaveTrendDraft(l.ctx, &trendrpc.SaveTrendDraftRequest{
		Draft: apiDraftToRPC(req.Draft, uint64(uid)),
	})
	if err != nil {
		return nil, err
	}

	return &types.SaveTrendDraftResponse{Draft: rpcDraftToAPI(res.Draft)}, nil
}

func apiDraftToRPC(d *types.TrendDraft, userId uint64) *trendrpc.TrendDraft {
	return &trendrpc.TrendDraft{
		Id:           d.Id,
		UserId:       userId,
		Type:         d.Type,
		Content:      d.Content,
		Title:        d.Title,
		Resources:    d.Resources,
		CoverUrl:     d.CoverUrl,
		ShareUrl:     d.ShareUrl,
		PositionName: d.PositionName,
		Longitude:    d.Longitude,
		Latitude:     d.Latitude,
		Scope:        d.Scope,
		OpenReply:    d.OpenReply,
		State:        d.State,
	}
}

func rpcDraftToAPI(d *trendrpc.TrendDraft) *types.TrendDraft {
	if d == nil {
		return nil
	}
	return &types.TrendDraft{
		Id:           d.Id,
		Type:         d.Type,
		Content:      d.Content,
		Title:        d.Title,
		Resources:    d.Resources,
		CoverUrl:     d.CoverUrl,
		ShareUrl:     d.ShareUrl,
		PositionName: d.PositionName,
		Longitude:    d.Longitude,
		Latitude:     d.Latitude,
		Scope:        d.Scope,
		OpenReply:    d.OpenReply,
		State:        d.State,
		CreateTime:   d.CreateTime,
		UpdateTime:   d.UpdateTime,
	}
}
