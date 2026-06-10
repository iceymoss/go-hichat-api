package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/notify"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/pkg/constants"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type GetTrendMessageUnreadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendMessageUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendMessageUnreadLogic {
	return &GetTrendMessageUnreadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTrendMessageUnread 获取动态消息未读数（总数 + 按类型明细）
func (l *GetTrendMessageUnreadLogic) GetTrendMessageUnread(in *trend.GetTrendMessageUnreadReq) (*trend.GetTrendMessageUnreadResp, error) {
	byType, err := l.svcCtx.TrendMessage.CountUnreadByType(l.ctx, in.UserId)
	if err != nil {
		zLog.Error("GetTrendMessageUnread: count failed", zap.Uint64("userId", in.UserId), zap.Error(err))
		return nil, err
	}

	return &trend.GetTrendMessageUnreadResp{
		Total:     notify.SumUnread(byType),
		Like:      byType[uint64(constants.TrendMsgLike)],
		Comment:   byType[uint64(constants.TrendMsgComment)],
		Reply:     byType[uint64(constants.TrendMsgReply)],
		AtTrend:   byType[uint64(constants.TrendMsgAtTrend)],
		AtComment: byType[uint64(constants.TrendMsgAtComment)],
	}, nil
}
