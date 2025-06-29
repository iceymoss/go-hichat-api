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

type MarkLikesReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMarkLikesReadLogic 标记点赞为已读
func NewMarkLikesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkLikesReadLogic {
	return &MarkLikesReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkLikesReadLogic) MarkLikesRead(req *types.MarkLikesReadRequest) (resp *types.MarkLikesReadResponse, err error) {
	uid := utils.GetUser(l.ctx)
	_, err = l.svcCtx.Trend.MarkLikesRead(l.ctx, &trend.MarkLikesReadRequest{
		UserId:  uint32(uid),
		LikeIds: req.LikeIDs,
	})
	if err != nil {
		zLog.Error("标记点赞已读失败", zap.Any("req", req), zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}

	resp = &types.MarkLikesReadResponse{}

	return
}
