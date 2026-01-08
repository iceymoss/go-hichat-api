package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupQuitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupQuitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupQuitLogic {
	return &GroupQuitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupQuitLogic) GroupQuit(req *types.GroupQuitReq) (resp *types.GroupQuitResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.GroupQuit(l.ctx, &social.GroupQuitReq{
		UserId:  uid,
		GroupId: req.GroupId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupQuitResp{}, nil
}
