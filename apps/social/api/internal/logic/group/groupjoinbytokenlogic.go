package group

import (
	"context"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupJoinByTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupJoinByTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupJoinByTokenLogic {
	return &GroupJoinByTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupJoinByTokenLogic) GroupJoinByToken(req *types.GroupJoinByTokenReq) (resp *types.GroupJoinByTokenResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GroupJoinByToken(l.ctx, &social.GroupJoinByTokenReq{
		UserId: uid,
		Token:  req.Token,
		ReqMsg: req.ReqMsg,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupJoinByTokenResp{
		GroupId: rpcResp.GroupId,
		IsPass:  rpcResp.IsPass,
	}, nil
}
