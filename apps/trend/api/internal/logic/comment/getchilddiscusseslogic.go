package comment

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChildDiscussesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取子评论列表
func NewGetChildDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildDiscussesLogic {
	return &GetChildDiscussesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChildDiscussesLogic) GetChildDiscusses(req *types.GetChildDiscussesReq) (resp *types.DiscussesListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
