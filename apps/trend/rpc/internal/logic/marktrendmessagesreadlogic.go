package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type MarkTrendMessagesReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkTrendMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkTrendMessagesReadLogic {
	return &MarkTrendMessagesReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkTrendMessagesRead 全部标记为已读
func (l *MarkTrendMessagesReadLogic) MarkTrendMessagesRead(in *trend.MarkTrendMessagesReadReq) (*trend.MarkTrendMessagesReadResp, error) {
	if err := l.svcCtx.TrendMessage.MarkAllRead(l.ctx, in.UserId); err != nil {
		zLog.Error("MarkTrendMessagesRead: mark all read failed", zap.Uint64("userId", in.UserId), zap.Error(err))
		return nil, err
	}
	return &trend.MarkTrendMessagesReadResp{}, nil
}
