package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupAnnouncementCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupAnnouncementCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupAnnouncementCreateLogic {
	return &GroupAnnouncementCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupAnnouncementCreateLogic) GroupAnnouncementCreate(req *types.GroupAnnouncementCreateReq) (resp *types.GroupAnnouncementCreateResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GroupAnnouncementCreate(l.ctx, &social.GroupAnnouncementCreateReq{
		UserId:  uid,
		GroupId: req.GroupId,
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupAnnouncementCreateResp{Id: rpcResp.Id}, nil
}
