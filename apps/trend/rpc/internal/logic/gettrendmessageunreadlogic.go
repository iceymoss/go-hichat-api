package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
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

// 获取动态消息未读数（总数 + 按类型明细）
func (l *GetTrendMessageUnreadLogic) GetTrendMessageUnread(in *trend.GetTrendMessageUnreadReq) (*trend.GetTrendMessageUnreadResp, error) {
	// todo: add your logic here and delete this line

	return &trend.GetTrendMessageUnreadResp{}, nil
}
