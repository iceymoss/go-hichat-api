package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupTransferOwnerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 转让群主（仅群主）
func NewGroupTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupTransferOwnerLogic {
	return &GroupTransferOwnerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupTransferOwnerLogic) GroupTransferOwner(req *types.GroupTransferOwnerReq) (resp *types.GroupTransferOwnerResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	keepOld := true
	if req.KeepOldOwnerAsAdmin != nil {
		keepOld = *req.KeepOldOwnerAsAdmin
	}

	_, err = l.svcCtx.Social.GroupTransferOwner(l.ctx, &social.GroupTransferOwnerReq{
		UserId:              uid,
		GroupId:             req.GroupId,
		NewOwnerId:          req.NewOwnerId,
		KeepOldOwnerAsAdmin: keepOld,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupTransferOwnerResp{}, nil
}
