package group

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupPutInHandleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GroupPutInHandleLogic {
	return &GroupPutInHandleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *GroupPutInHandleLogic) GroupPutInHandle(req *types.GroupPutInHandleReq) (*types.GroupPutInHandleResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if id, err := strconv.ParseUint(uid, 10, 64); err != nil || id == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	requestID, err := parseGroupID(req.RequestId, "group request id")
	if err != nil {
		return nil, err
	}
	if req.HandleResult != 1 && req.HandleResult != 2 {
		return nil, status.Error(codes.InvalidArgument, "handle result must be 1 or 2")
	}
	res, err := l.svcCtx.Social.GroupPutInHandle(l.ctx, &social.GroupPutInHandleReq{
		RequestId: requestID, HandleUid: uid, HandleResult: req.HandleResult,
		ActorUid: uid, HandleMsg: req.HandleMsg,
	})
	if err != nil {
		return nil, err
	}
	return &types.GroupPutInHandleResp{RequestId: strconv.FormatUint(res.RequestId, 10), HandleResult: res.HandleResult, Idempotent: res.Idempotent}, nil
}
