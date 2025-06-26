package comment

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDiscussesTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取评论树
func NewGetDiscussesTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiscussesTreeLogic {
	return &GetDiscussesTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDiscussesTreeLogic) GetDiscussesTree(req *types.GetDiscussesReq) (resp *types.DiscussesTreeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
