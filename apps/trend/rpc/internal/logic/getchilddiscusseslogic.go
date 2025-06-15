package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChildDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChildDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildDiscussesLogic {
	return &GetChildDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 5. 获取子评论（指定父评论下的所有评论）
func (l *GetChildDiscussesLogic) GetChildDiscusses(in *trend.GetChildDiscussesReq) (*trend.DiscussesListResp, error) {
	// todo: add your logic here and delete this line

	return &trend.DiscussesListResp{}, nil
}
