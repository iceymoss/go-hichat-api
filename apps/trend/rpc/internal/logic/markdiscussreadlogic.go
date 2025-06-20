package logic

import (
	"context"

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

// MarkDiscussRead 8. 标记评论为已读
func (l *MarkDiscussReadLogic) MarkDiscussRead(in *trend.MarkDiscussRequest) (*trend.MarkDiscussResponse, error) {
	// todo: add your logic here and delete this line

	return &trend.MarkDiscussResponse{}, nil
}
