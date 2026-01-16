package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupAnnouncementPinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupAnnouncementPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAnnouncementPinLogic {
	return &GroupAnnouncementPinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupAnnouncementPinLogic) GroupAnnouncementPin(req *types.GroupAnnouncementPinReq) (resp *types.GroupAnnouncementPinResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.GroupAnnouncementPin(l.ctx, &social.GroupAnnouncementPinReq{
		UserId:         uid,
		GroupId:        req.GroupId,
		AnnouncementId: req.AnnouncementId,
		Pinned:         req.Pinned,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupAnnouncementPinResp{}, nil
}
