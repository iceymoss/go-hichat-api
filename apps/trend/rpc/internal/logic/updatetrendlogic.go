package logic

import (
	"context"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTrendLogic {
	return &UpdateTrendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateTrend 更新动态
func (l *UpdateTrendLogic) UpdateTrend(in *trend.UpdateTrendRequest) (*trend.UpdateTrendResponse, error) {
	// 更新置顶
	var isTop bool
	if in.IsTop == 1 {
		isTop = true
	}
	err := l.svcCtx.Trend.SetTop(l.ctx, uint64(in.TrendId), isTop)
	if err != nil {
		zLog.Error("设置动态置顶失败", zap.Error(err))
		return nil, err
	}

	// 更新评论区
	var isOpen bool
	if in.OpenReply == 1 {
		isOpen = true
	}
	err = l.svcCtx.Trend.OpenReply(l.ctx, uint64(in.TrendId), isOpen)
	if err != nil {
		zLog.Error("设置动态评论区失败", zap.Error(err))
		return nil, err
	}

	// 更新可见范围(1-仅自己,2-仅好友,3-所有人)。scope 为 0 表示本次不修改范围
	scope := int(in.Scope)
	if scope >= 1 && scope <= 3 {
		circleState := 1 // 仅好友/所有人 进朋友圈流
		if scope == 1 {  // 仅自己
			circleState = 2
		}
		err = l.svcCtx.Trend.SetScope(l.ctx, uint64(in.TrendId), scope, circleState)
		if err != nil {
			zLog.Error("设置动态权限失败", zap.Error(err))
			return nil, err
		}
	}

	return &trend.UpdateTrendResponse{}, nil
}
