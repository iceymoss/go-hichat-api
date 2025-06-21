package logic

import (
	"context"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkDiscussReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkDiscussReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkDiscussReadLogic {
	return &MarkDiscussReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkDiscussRead 标记评论为已读
func (l *MarkDiscussReadLogic) MarkDiscussRead(in *trend.MarkDiscussRequest) (*trend.MarkDiscussResponse, error) {
	// 更新未读状态,标记为已读
	err := l.svcCtx.TrendDiscuss.MarkReadById(l.ctx, uint64(in.UserId), in.DisIds)
	if err != nil {
		zLog.Error("GetUnreadReplies.MarkReadById: 标记为已读失败", zap.Any("idList", in.DisIds), zap.Error(err))
		return nil, err
	}

	return &trend.MarkDiscussResponse{}, nil
}
