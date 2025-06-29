package comment

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDiscussLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteDiscussLogic 删除评论
func NewDeleteDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDiscussLogic {
	return &DeleteDiscussLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDiscussLogic) DeleteDiscuss(req *types.DeleteDiscussReq) (resp *types.DeleteDiscussResp, err error) {
	_, err = l.svcCtx.Trend.DeleteDiscuss(l.ctx, &trend.DeleteDiscussReq{Id: req.ID})
	if err != nil {
		zLog.Error("删除评论失败", zap.Any("id", req.ID), zap.Error(err))
		return nil, err
	}

	resp = &types.DeleteDiscussResp{Success: true}

	return
}
