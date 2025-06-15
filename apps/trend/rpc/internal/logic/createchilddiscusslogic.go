package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChildDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChildDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChildDiscussLogic {
	return &CreateChildDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 2. 发表多级评论（二级、三级等）
func (l *CreateChildDiscussLogic) CreateChildDiscuss(in *trend.CreateDiscussReq) (*trend.CreateDiscussResp, error) {
	// todo: add your logic here and delete this line

	return &trend.CreateDiscussResp{}, nil
}
