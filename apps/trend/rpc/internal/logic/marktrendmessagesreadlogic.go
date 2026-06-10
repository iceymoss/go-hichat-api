package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
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

// 全部标记为已读
func (l *MarkTrendMessagesReadLogic) MarkTrendMessagesRead(in *trend.MarkTrendMessagesReadReq) (*trend.MarkTrendMessagesReadResp, error) {
	// todo: add your logic here and delete this line

	return &trend.MarkTrendMessagesReadResp{}, nil
}
