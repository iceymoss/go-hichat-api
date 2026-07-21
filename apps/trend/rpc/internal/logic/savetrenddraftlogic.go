package logic

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/iceymoss/go-hichat-api/apps/trend/models"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveTrendDraftLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveTrendDraftLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveTrendDraftLogic {
	return &SaveTrendDraftLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存动态草稿
func (l *SaveTrendDraftLogic) SaveTrendDraft(in *trend.SaveTrendDraftRequest) (*trend.SaveTrendDraftResponse, error) {
	if in.Draft == nil || in.Draft.UserId == 0 {
		return nil, errors.New("草稿用户不能为空")
	}

	resources, err := json.Marshal(in.Draft.Resources)
	if err != nil {
		return nil, err
	}
	openReply := in.Draft.OpenReply
	if openReply == 0 {
		openReply = 1
	}
	scope := in.Draft.Scope
	if scope < 1 || scope > 3 {
		scope = 2
	}
	typeVal := in.Draft.Type
	if typeVal == 0 {
		typeVal = 1
	}

	saved, err := l.svcCtx.TrendDraft.Save(l.ctx, &models.TrendDraft{
		Id:           in.Draft.Id,
		UserId:       in.Draft.UserId,
		Type:         int64(typeVal),
		Content:      in.Draft.Content,
		Title:        in.Draft.Title,
		Resources:    string(resources),
		CoverUrl:     in.Draft.CoverUrl,
		ShareUrl:     in.Draft.ShareUrl,
		PositionName: in.Draft.PositionName,
		Longitude:    in.Draft.Longitude,
		Latitude:     in.Draft.Latitude,
		Scope:        int64(scope),
		OpenReply:    int64(openReply),
		State:        int64(in.Draft.State),
	})
	if err != nil {
		return nil, err
	}

	return &trend.SaveTrendDraftResponse{Draft: draftToResp(saved)}, nil
}

func draftToResp(v *models.TrendDraft) *trend.TrendDraft {
	if v == nil {
		return nil
	}
	var resources []string
	_ = json.Unmarshal([]byte(v.Resources), &resources)
	return &trend.TrendDraft{
		Id:           v.Id,
		UserId:       v.UserId,
		Type:         int32(v.Type),
		Content:      v.Content,
		Title:        v.Title,
		Resources:    resources,
		CoverUrl:     v.CoverUrl,
		ShareUrl:     v.ShareUrl,
		PositionName: v.PositionName,
		Longitude:    v.Longitude,
		Latitude:     v.Latitude,
		Scope:        int32(v.Scope),
		OpenReply:    int32(v.OpenReply),
		State:        int32(v.State),
		CreateTime:   v.CreatedAt.Unix(),
		UpdateTime:   v.UpdatedAt.Unix(),
	}
}
