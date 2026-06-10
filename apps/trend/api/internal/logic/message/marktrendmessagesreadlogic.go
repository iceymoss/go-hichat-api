package message

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"
	"github.com/iceymoss/go-hichat-api/apps/user/utils"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type MarkTrendMessagesReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMarkTrendMessagesReadLogic 动态消息全部标记为已读
func NewMarkTrendMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkTrendMessagesReadLogic {
	return &MarkTrendMessagesReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkTrendMessagesReadLogic) MarkTrendMessagesRead() (resp *types.MarkTrendMessagesReadResp, err error) {
	uid := utils.GetUser(l.ctx)
	if _, err = l.svcCtx.Trend.MarkTrendMessagesRead(l.ctx, &trend.MarkTrendMessagesReadReq{
		UserId: uint64(uid),
	}); err != nil {
		zLog.Error("MarkTrendMessagesRead: rpc failed", zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}
	return &types.MarkTrendMessagesReadResp{}, nil
}
