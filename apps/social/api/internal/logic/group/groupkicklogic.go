package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupKickLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupKickLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupKickLogic {
	return &GroupKickLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupKickLogic) GroupKick(req *types.GroupKickReq) (resp *types.GroupKickResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.GroupKick(l.ctx, &social.GroupKickReq{
		UserId:    uid,
		GroupId:   req.GroupId,
		MemberIds: req.MemberIds,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupKickResp{}, nil
}
