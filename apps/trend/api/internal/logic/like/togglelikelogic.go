package like

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ToggleLikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewToggleLikeLogic 点赞/取消点赞
func NewToggleLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ToggleLikeLogic {
	return &ToggleLikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ToggleLikeLogic) ToggleLike(req *types.LikeToggleRequest) (resp *types.LikeToggleResponse, err error) {
	uid := utils.GetUser(l.ctx)
	_, err = l.svcCtx.Trend.ToggleLike(l.ctx, &trend.LikeToggleRequest{
		UserId:   uint32(uid),
		TrendId:  req.TrendID,
		AuthorId: req.AuthorID,
		LikeType: int32(req.LikeType),
	})
	if err != nil {
		zLog.Error("点赞失败", zap.Any("uid", uid), zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	resp = &types.LikeToggleResponse{}

	return
}
