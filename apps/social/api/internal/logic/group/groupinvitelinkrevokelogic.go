package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLinkRevokeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInviteLinkRevokeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLinkRevokeLogic {
	return &GroupInviteLinkRevokeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupInviteLinkRevokeLogic) GroupInviteLinkRevoke(req *types.GroupInviteLinkRevokeReq) (resp *types.GroupInviteLinkRevokeResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.GroupInviteLinkRevoke(l.ctx, &social.GroupInviteLinkRevokeReq{
		UserId:  uid,
		GroupId: req.GroupId,
		Token:   req.Token,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupInviteLinkRevokeResp{}, nil
}
