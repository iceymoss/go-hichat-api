package friend

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FriendPutInHandleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 好友申请处理
func NewFriendPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInHandleLogic {
	return &FriendPutInHandleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInHandleLogic) FriendPutInHandle(req *types.FriendPutInHandleReq) (resp *types.FriendPutInHandleResp, err error) {
	curUid := ctxdata.GetUId(l.ctx)
	uid, parseErr := strconv.ParseUint(curUid, 10, 64)
	if parseErr != nil || uid == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	if req.FriendReqId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "friend request id must be positive")
	}
	if req.HandleResult != 1 && req.HandleResult != 2 {
		return nil, status.Error(codes.InvalidArgument, "handle result must be 1 or 2")
	}
	rpcResp, err := l.svcCtx.Social.FriendPutInHandle(l.ctx, &social.FriendPutInHandleReq{
		FriendReqId:  req.FriendReqId,
		UserId:       curUid,
		ActorUid:     curUid,
		HandleResult: req.HandleResult,
		Remark:       req.Remark,
		HandleMsg:    req.HandleMsg,
		Tags:         req.Tags,
	})
	if err != nil {
		zLog.Error("friend req handle err", zap.Error(err))
		return nil, err
	}

	return &types.FriendPutInHandleResp{
		RequestId: rpcResp.RequestId, HandleResult: rpcResp.HandleResult, Idempotent: rpcResp.Idempotent,
	}, nil
}
