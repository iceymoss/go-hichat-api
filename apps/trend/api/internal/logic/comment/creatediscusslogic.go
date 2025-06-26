package comment

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDiscussLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建新评论
func NewCreateDiscussLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDiscussLogic {
	return &CreateDiscussLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDiscussLogic) CreateDiscuss(req *types.CreateDiscussReq) (resp *types.CreateDiscussResp, err error) {
	// todo: add your logic here and delete this line

	return
}
