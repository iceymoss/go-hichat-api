package trend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/pkg/systemconfig"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendPublishConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取动态发布配置
func NewGetTrendPublishConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendPublishConfigLogic {
	return &GetTrendPublishConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTrendPublishConfigLogic) GetTrendPublishConfig() (resp *types.TrendPublishConfigResponse, err error) {
	cfg := defaultTrendPublishConfig()
	cfg.MaxImageCount = systemconfig.GetInt(l.ctx, "trendMaxImageCount", cfg.MaxImageCount)
	cfg.MaxImageSizeMB = systemconfig.GetInt(l.ctx, "trendMaxImageSizeMB", cfg.MaxImageSizeMB)
	cfg.AllowedImageTypes = systemconfig.GetStringList(l.ctx, "trendAllowedImageTypes", cfg.AllowedImageTypes)
	cfg.MaxVideoCount = systemconfig.GetInt(l.ctx, "trendMaxVideoCount", cfg.MaxVideoCount)
	cfg.MaxVideoSizeMB = systemconfig.GetInt(l.ctx, "trendMaxVideoSizeMB", cfg.MaxVideoSizeMB)
	cfg.MaxVideoDurationSec = systemconfig.GetInt(l.ctx, "trendMaxVideoDurationSec", cfg.MaxVideoDurationSec)
	cfg.AllowedVideoTypes = systemconfig.GetStringList(l.ctx, "trendAllowedVideoTypes", cfg.AllowedVideoTypes)
	cfg.ImageCompressionEnable = systemconfig.GetBool(l.ctx, "trendImageCompressionEnabled", cfg.ImageCompressionEnable)
	cfg.VideoCompressionEnable = systemconfig.GetBool(l.ctx, "trendVideoCompressionEnabled", cfg.VideoCompressionEnable)
	cfg.MediaReviewEnable = systemconfig.GetBool(l.ctx, "trendMediaReviewEnabled", cfg.MediaReviewEnable)
	return cfg, nil
}

func defaultTrendPublishConfig() *types.TrendPublishConfigResponse {
	return &types.TrendPublishConfigResponse{
		MaxImageCount:          9,
		MaxImageSizeMB:         50,
		AllowedImageTypes:      []string{"jpg", "jpeg", "png", "webp", "gif"},
		MaxVideoCount:          1,
		MaxVideoSizeMB:         500,
		MaxVideoDurationSec:    0,
		AllowedVideoTypes:      []string{"mp4", "mov", "webm"},
		ImageCompressionEnable: true,
		VideoCompressionEnable: true,
		MediaReviewEnable:      false,
	}
}
