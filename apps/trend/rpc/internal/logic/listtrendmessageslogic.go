package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTrendMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTrendMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTrendMessagesLogic {
	return &ListTrendMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ===== 动态消息通知 =====
func (l *ListTrendMessagesLogic) ListTrendMessages(in *trend.ListTrendMessagesReq) (*trend.ListTrendMessagesResp, error) {
	// todo: add your logic here and delete this line

	return &trend.ListTrendMessagesResp{}, nil
}
