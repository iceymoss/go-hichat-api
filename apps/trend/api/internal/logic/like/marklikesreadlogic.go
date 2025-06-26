package like

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkLikesReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记点赞为已读
func NewMarkLikesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkLikesReadLogic {
	return &MarkLikesReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkLikesReadLogic) MarkLikesRead(req *types.MarkLikesReadRequest) (resp *types.MarkLikesReadResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
