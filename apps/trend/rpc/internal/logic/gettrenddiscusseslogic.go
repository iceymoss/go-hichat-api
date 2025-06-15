package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/rpc/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTrendDiscussesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTrendDiscussesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTrendDiscussesLogic {
	return &GetTrendDiscussesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 3. 获取动态的多级评论树（树形结构）
func (l *GetTrendDiscussesLogic) GetTrendDiscusses(in *trend.GetDiscussesReq) (*trend.DiscussesTreeResp, error) {
	// todo: add your logic here and delete this line

	return &trend.DiscussesTreeResp{}, nil
}
