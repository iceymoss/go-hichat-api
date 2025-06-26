package comment

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkDiscussReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记评论为已读
func NewMarkDiscussReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkDiscussReadLogic {
	return &MarkDiscussReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkDiscussReadLogic) MarkDiscussRead(req *types.MarkDiscussRequest) (resp *types.MarkDiscussResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
