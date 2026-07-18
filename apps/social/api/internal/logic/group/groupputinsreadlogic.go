package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupPutInsReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGroupPutInsReadLogic 把我管理的群收到的入群申请全部标记已读
func NewGroupPutInsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInsReadLogic {
	return &GroupPutInsReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupPutInsReadLogic) GroupPutInsRead(req *types.GroupPutInsReadReq) (resp *types.GroupPutInsReadResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.MarkGroupReqRead(l.ctx, &socialclient.MarkGroupReqReadReq{
		UserId:     uid,
		RequestIds: req.RequestIds,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupPutInsReadResp{Count: rpcResp.Count, Apply: rpcResp.Apply, Result: rpcResp.Result, Invite: rpcResp.Invite}, nil
}
