package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupInviteLinkListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupInviteLinkListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupInviteLinkListLogic {
	return &GroupInviteLinkListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupInviteLinkListLogic) GroupInviteLinkList(req *types.GroupInviteLinkListReq) (resp *types.GroupInviteLinkListResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GroupInviteLinkList(l.ctx, &social.GroupInviteLinkListReq{
		UserId:         uid,
		GroupId:        req.GroupId,
		IncludeRevoked: req.IncludeRevoked,
	})
	if err != nil {
		return nil, err
	}

	list := make([]types.GroupInviteLink, 0, len(rpcResp.List))
	for _, it := range rpcResp.List {
		list = append(list, types.GroupInviteLink{
			Token:     it.Token,
			GroupId:   it.GroupId,
			CreatedBy: it.CreatedBy,
			CreatedAt: it.CreatedAt,
			ExpireAt:  it.ExpireAt,
			MaxUses:   it.MaxUses,
			UsedCount: it.UsedCount,
			Revoked:   it.Revoked,
			RevokedAt: it.RevokedAt,
		})
	}

	return &types.GroupInviteLinkListResp{List: list}, nil
}
