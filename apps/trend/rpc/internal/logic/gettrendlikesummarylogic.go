package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

// GetTrendLikeSummary 获取动态点赞摘要 (用于动态下方展示)
func (l *GetTrendLikeSummaryLogic) GetTrendLikeSummary(in *trend.GetTrendLikeSummaryRequest) (*trend.GetTrendLikeSummaryResponse, error) {
	trendInfo, err := l.svcCtx.Trend.FindOne(l.ctx, uint64(in.TrendId), []string{"id", "agree_count"})
	if err != nil {
		zLog.Error("get trend failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, err
	}

	// 查询当前用户是否点赞
	curUserAgree, err := l.svcCtx.TrendAgree.FindOneByUseridTrendId(l.ctx, uint64(in.UserId), uint64(in.TrendId), 1)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		zLog.Error("get trend agree failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, err
	}

	// 获取点赞数据
	agreeList, count, err := l.svcCtx.TrendAgree.GetAgreeByTrendId(l.ctx, uint64(in.TrendId), []string{"*"}, 0, 3)
	if err != nil {
		zLog.Error("get trend agree failed", zap.Any("trendId", in.TrendId), zap.Error(err))
		return nil, err
	}

	previewLikes := make([]*trend.LikeInfo, 0, len(agreeList))
	for _, v := range agreeList {
		previewLikes = append(previewLikes, &trend.LikeInfo{
			Id:         v.Id,
			UserId:     uint32(v.Userid),
			TrendId:    uint32(v.TrendId),
			CreateTime: v.OpTime.Unix(),
		})
	}

	return &trend.GetTrendLikeSummaryResponse{Summary: &trend.TrendLikeSummary{
		TrendId:      in.TrendId,
		LikeCount:    uint32(trendInfo.AgreeCount),
		IsLiked:      curUserAgree != nil,
		PreviewLikes: previewLikes,
		Count:        uint32(count),
	}}, nil
}
