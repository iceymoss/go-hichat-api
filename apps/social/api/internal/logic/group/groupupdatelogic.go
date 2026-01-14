package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupUpdateLogic {
	return &GroupUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupUpdateLogic) GroupUpdate(req *types.GroupUpdateReq) (resp *types.GroupUpdateResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	isVerify := int32(-1)
	if req.IsVerify != nil {
		isVerify = *req.IsVerify
	}
	_, err = l.svcCtx.Social.GroupUpdate(l.ctx, &social.GroupUpdateReq{
		UserId:       uid,
		GroupId:      req.GroupId,
		Name:         req.Name,
		Icon:         req.Icon,
		Notification: req.Notification,
		IsVerify:     isVerify,
	})
	if err != nil {
		return nil, err
	}

	return &types.GroupUpdateResp{}, nil
}
