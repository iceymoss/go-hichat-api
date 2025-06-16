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
	discus, err := l.svcCtx.TrendDiscuss.FindOne(l.ctx, in.Id)
	if err != nil {
		zLog.Error("DeleteDiscuss.FindOne: get filed", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}
	err = l.svcCtx.TrendDiscuss.Delete(l.ctx, in.Id)
	if err != nil {
		zLog.Error("DeleteDiscuss.Delete: delete filed", zap.Any("id", in.Id), zap.Error(err))
		return nil, err
	}

	// 删除关联的评论
	//if discus.Level == 1 {
	//	err := l.svcCtx.TrendDiscuss.DeleteByRoots(l.ctx, []uint64{discus.Rootid})
	//	if err != nil {
	//		zLog.Error("DeleteDiscuss.v: delete filed", zap.Any("id", in.Id), zap.Error(err))
	//		// 不中断
	//	}
	//}

	// 动态总评论数-1
	// 更新动态总评论数
	err = l.svcCtx.Trend.IncAgreeOrReply(l.ctx, uint64(discus.Trendid), 1, -1)
	if err != nil {
		zLog.Error("DeleteDiscuss.IncAgreeOrReply: -inc reply failed", zap.Error(err))
		// 不中断处理，业务上可以接受丢失
	}

	return &trend.DeleteDiscussResp{
		Success: true,
	}, nil
}
