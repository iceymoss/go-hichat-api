package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLogic {
	return &GroupInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupInviteLogic) GroupInvite(req *types.GroupInviteReq) (resp *types.GroupInviteResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.GroupInvite(l.ctx, &social.GroupInviteReq{
		UserId:    uid,
		GroupId:   req.GroupId,
		FriendIds: req.FriendIds,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupInviteResp{}, nil
}
