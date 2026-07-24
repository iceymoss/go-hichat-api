package group

import (
	"context"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupPutInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInLogic {
	return &GroupPutInLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GroupPutInLogic) GroupPutIn(req *types.GroupPutInReq) (*types.GroupPutInResp, error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}

	// Legacy identity/source fields are intentionally ignored for public direct applications.
	res, err := l.svcCtx.Social.GroupPutin(l.ctx, &social.GroupPutinReq{
		GroupId: req.GroupId, ReqId: uid, ReqMsg: req.ReqMsg, ReqTime: time.Now().Unix(),
		JoinSource: 1, ActorUid: uid,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupPutInResp{
		GroupId: res.GroupIdString, IsPass: int(res.IsPass), RequestId: formatOptionalID(res.RequestId),
		Status: res.Status, AlreadyPending: res.AlreadyPending, AlreadyMember: res.AlreadyMember,
	}, nil
}
