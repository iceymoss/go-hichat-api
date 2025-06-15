package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type DeleteDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDiscussLogic {
	return &DeleteDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteDiscuss 删除评论（支持任意级别评论删除）
func (l *DeleteDiscussLogic) DeleteDiscuss(in *trend.DeleteDiscussReq) (*trend.DeleteDiscussResp, error) {
	err := l.svcCtx.TrendDiscuss.Delete(l.ctx, in.Id)
	if err != nil {
		zLog.Error("DeleteDiscuss.Delete: delete filed", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	return &trend.DeleteDiscussResp{}, nil
}
