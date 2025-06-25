package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

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

// BatchGetTrendLikeSummary 批量获取点赞摘要 (用于动态流)
func (l *BatchGetTrendLikeSummaryLogic) BatchGetTrendLikeSummary(in *trend.BatchTrendLikeSummaryRequest) (*trend.BatchTrendLikeSummaryResponse, error) {
	// 批量获取动态点赞数据
	agreeBindTrend, err := l.svcCtx.TrendAgree.GetTrendAgree(l.ctx, in.TrendIds, 1)
	if err != nil {
		zLog.Error("BatchGetTrendLikeSummary.GetTrendAgree: get trend agree list failed")
		return nil, err
	}

	list := make(map[uint32][]*trend.LikeInfo, len(agreeBindTrend))
	curUserAgree := make(map[uint64]bool, len(agreeBindTrend))
	for k, v := range agreeBindTrend {
		if _, ok := list[uint32(k)]; !ok {
			list[uint32(k)] = make([]*trend.LikeInfo, 0, len(v))
			continue
		}
		for _, agreeItem := range v {
			list[uint32(k)] = append(list[uint32(k)], &trend.LikeInfo{
				Id:         agreeItem.Id,
				UserId:     uint32(agreeItem.Userid),
				TrendId:    uint32(agreeItem.TrendId),
				CreateTime: agreeItem.CreateTime.Unix(),
			})
			if agreeItem.Userid == uint64(in.UserId) {
				curUserAgree[agreeItem.TrendId] = true
			}
		}
	}

	summary := make(map[uint32]*trend.TrendLikeSummary, len(agreeBindTrend))
	for trendId := range agreeBindTrend {
		summary[uint32(trendId)] = &trend.TrendLikeSummary{
			TrendId:      uint32(trendId),
			LikeCount:    uint32(len(list[uint32(trendId)])),
			IsLiked:      curUserAgree[trendId],
			PreviewLikes: list[uint32(trendId)],
			Count:        0,
		}
	}

	return &trend.BatchTrendLikeSummaryResponse{
		Summaries: summary,
	}, nil
}
