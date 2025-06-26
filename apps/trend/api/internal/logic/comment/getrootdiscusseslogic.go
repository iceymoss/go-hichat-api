package comment

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRootDiscussesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取一级评论
func NewGetRootDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootDiscussesLogic {
	return &GetRootDiscussesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRootDiscussesLogic) GetRootDiscusses(req *types.GetDiscussesReq) (resp *types.DiscussesListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
