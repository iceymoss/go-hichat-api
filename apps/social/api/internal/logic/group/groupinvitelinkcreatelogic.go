package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLinkCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInviteLinkCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLinkCreateLogic {
	return &GroupInviteLinkCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupInviteLinkCreateLogic) GroupInviteLinkCreate(req *types.GroupInviteLinkCreateReq) (resp *types.GroupInviteLinkCreateResp, err error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}
	rpcResp, err := l.svcCtx.Social.GroupInviteLinkCreate(l.ctx, &social.GroupInviteLinkCreateReq{
		UserId:        uid,
		ActorUid:      uid,
		GroupId:       req.GroupId,
		ExpireSeconds: req.ExpireSeconds,
		MaxUses:       req.MaxUses,
	})
	if err != nil {
		return nil, err
	}

	var link types.GroupInviteLink
	if rpcResp.Link != nil {
		link = types.GroupInviteLink{
			Token:     rpcResp.Link.Token,
			GroupId:   rpcResp.Link.GroupId,
			CreatedBy: rpcResp.Link.CreatedBy,
			CreatedAt: rpcResp.Link.CreatedAt,
			ExpireAt:  rpcResp.Link.ExpireAt,
			MaxUses:   rpcResp.Link.MaxUses,
			UsedCount: rpcResp.Link.UsedCount,
			Revoked:   rpcResp.Link.Revoked,
			RevokedAt: rpcResp.Link.RevokedAt,
		}
	}

	return &types.GroupInviteLinkCreateResp{Link: link}, nil
}
