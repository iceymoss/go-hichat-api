package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRootDiscussLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRootDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRootDiscussLogic {
	return &CreateRootDiscussLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 评论服务定义
func (l *CreateRootDiscussLogic) CreateRootDiscuss(in *trend.CreateDiscussReq) (*trend.CreateDiscussResp, error) {
	// todo: add your logic here and delete this line

	return &trend.CreateDiscussResp{}, nil
}
