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

type GetTrendMessageUnreadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetTrendMessageUnreadLogic 获取动态消息未读数（总数+按类型明细）
func NewGetTrendMessageUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendMessageUnreadLogic {
	return &GetTrendMessageUnreadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTrendMessageUnreadLogic) GetTrendMessageUnread() (resp *types.GetTrendMessageUnreadResp, err error) {
	uid := utils.GetUser(l.ctx)
	rpcResp, err := l.svcCtx.Trend.GetTrendMessageUnread(l.ctx, &trend.GetTrendMessageUnreadReq{
		UserId: uint64(uid),
	})
	if err != nil {
		zLog.Error("GetTrendMessageUnread: rpc failed", zap.Any("uid", uid), zap.Error(err))
		return nil, err
	}

	return &types.GetTrendMessageUnreadResp{
		Total:     rpcResp.Total,
		Like:      rpcResp.Like,
		Comment:   rpcResp.Comment,
		Reply:     rpcResp.Reply,
		AtTrend:   rpcResp.AtTrend,
		AtComment: rpcResp.AtComment,
	}, nil
}
