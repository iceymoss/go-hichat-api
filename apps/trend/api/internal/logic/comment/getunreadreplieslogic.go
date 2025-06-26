package comment

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadRepliesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取未读回复通知
func NewGetUnreadRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadRepliesLogic {
	return &GetUnreadRepliesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUnreadRepliesLogic) GetUnreadReplies(req *types.GetUnreadRepliesReq) (resp *types.RepliesListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
