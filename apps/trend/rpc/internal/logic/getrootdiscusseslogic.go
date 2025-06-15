package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRootDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRootDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootDiscussesLogic {
	return &GetRootDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 4. 获取一级评论（分页）
func (l *GetRootDiscussesLogic) GetRootDiscusses(in *trend.GetDiscussesReq) (*trend.DiscussesListResp, error) {
	// todo: add your logic here and delete this line

	return &trend.DiscussesListResp{}, nil
}
