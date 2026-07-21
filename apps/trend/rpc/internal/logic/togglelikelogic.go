package logic

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/notify"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/transaction"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type ToggleLikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewToggleLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ToggleLikeLogic {
	return &ToggleLikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ToggleLike 点赞/取消点赞
func (l *ToggleLikeLogic) ToggleLike(in *trend.LikeToggleRequest) (*trend.LikeToggleResponse, error) {
	// delta: 1=本次新增点赞, -1=本次取消, 0=无变化（幂等）
	var delta int
	// 实例化事务管理器
	tx := transaction.NewManager()
	txErr := tx.Execute(l.ctx, nil, func(ctx context.Context) error {
		var err error
		delta, err = l.svcCtx.TrendAgree.AgreeInc(l.ctx, uint64(in.UserId), uint64(in.AuthorId), uint64(in.TrendId), int(in.LikeType))
		if err != nil {
			return fmt.Errorf("用户: %d 点赞动态: %d 失败: %s", in.UserId, in.TrendId, err)
		}

		// 仅在点赞状态真正变化时调整动态点赞总数(幂等,避免重复点击导致计数漂移)
		if delta != 0 {
			if err = l.svcCtx.Trend.IncAgreeOrReply(l.ctx, uint64(in.TrendId), 0, delta); err != nil {
				return fmt.Errorf("更新动态点赞总数失败: trend_id: %d error: %s", in.TrendId, err)
			}
		}

		return nil
	})

	if txErr != nil {
		zLog.Error("ToggleLike.Execute: 执行事务失败", zap.Any("trendId", in.TrendId), zap.Error(txErr))
		return nil, txErr
	}

	resp := &trend.LikeToggleResponse{Success: true}
	switch delta {
	case 1: // 新增点赞：生成点赞消息（自我点赞由 builder 过滤）
		resp.CreatedMessages = persistTrendMessages(l.ctx, l.svcCtx, notify.TrendNotifyEvent{
			Type:     constants.TrendMsgLike,
			ActorId:  uint64(in.UserId),
			TrendId:  uint64(in.TrendId),
			AuthorId: uint64(in.AuthorId),
		})
	case -1: // 取消点赞：级联软删对应点赞消息
		if err := l.svcCtx.TrendMessage.SoftDeleteLike(l.ctx, uint64(in.TrendId), uint64(in.UserId)); err != nil {
			zLog.Error("ToggleLike: 级联软删点赞消息失败", zap.Any("trendId", in.TrendId), zap.Any("uid", in.UserId), zap.Error(err))
		}
	}

	return resp, nil
}
