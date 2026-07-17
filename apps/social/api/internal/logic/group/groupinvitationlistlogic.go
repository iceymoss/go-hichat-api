package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInvitationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInvitationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInvitationListLogic {
	return &GroupInvitationListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GroupInvitationListLogic) GroupInvitationList(req *types.GroupInvitationListReq) (*types.GroupInvitationListResp, error) {
	uid, err := apiActor(l.ctx)
	if err != nil {
		return nil, err
	}
	filter := int32(-1)
	if req.Status != nil {
		filter = *req.Status
	}
	res, err := l.svcCtx.Social.GroupInvitationList(l.ctx, &social.GroupInvitationListReq{
		ActorUid: uid, Status: filter, Page: req.Page, Size: req.Size,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.GroupInvitation, len(res.List))
	for i := range res.List {
		list[i] = invitationType(res.List[i])
	}
	return &types.GroupInvitationListResp{List: list, Total: res.Total}, nil
}
