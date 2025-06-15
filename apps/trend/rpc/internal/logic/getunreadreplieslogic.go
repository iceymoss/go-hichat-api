package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadRepliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadRepliesLogic {
	return &GetUnreadRepliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 7. 获取未读回复
func (l *GetUnreadRepliesLogic) GetUnreadReplies(in *trend.GetUnreadRepliesReq) (*trend.RepliesListResp, error) {
	// todo: add your logic here and delete this line

	return &trend.RepliesListResp{}, nil
}
