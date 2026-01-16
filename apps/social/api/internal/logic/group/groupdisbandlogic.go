package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupDisbandLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解散群（仅群主）
func NewGroupDisbandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupDisbandLogic {
	return &GroupDisbandLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupDisbandLogic) GroupDisband(req *types.GroupDisbandReq) (resp *types.GroupDisbandResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.GroupDisband(l.ctx, &social.GroupDisbandReq{
		UserId:  uid,
		GroupId: req.GroupId,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupDisbandResp{}, nil
}
