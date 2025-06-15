package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDiscussLogic {
	return &DeleteDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 6. 删除评论（支持任意级别评论删除）
func (l *DeleteDiscussLogic) DeleteDiscuss(in *trend.DeleteDiscussReq) (*trend.DeleteDiscussResp, error) {
	// todo: add your logic here and delete this line

	return &trend.DeleteDiscussResp{}, nil
}
