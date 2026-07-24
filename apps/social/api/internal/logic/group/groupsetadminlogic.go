package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupSetAdminLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 设置/取消管理员（仅群主）
func NewGroupSetAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupSetAdminLogic {
	return &GroupSetAdminLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupSetAdminLogic) GroupSetAdmin(req *types.GroupSetAdminReq) (resp *types.GroupSetAdminResp, err error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.Social.GroupSetAdmin(l.ctx, &social.GroupSetAdminReq{
		UserId:    uid,
		ActorUid:  uid,
		GroupId:   req.GroupId,
		MemberIds: req.MemberIds,
		IsAdmin:   req.IsAdmin,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupSetAdminResp{}, nil
}
