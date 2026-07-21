package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/actor"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type FriendPutInDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除好友申请
func NewFriendPutInDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInDeleteLogic {
	return &FriendPutInDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInDeleteLogic) FriendPutInDelete(req *types.FriendPutInDeleteReq) (resp *types.FriendPutInDeleteResp, err error) {
	uid, err := actor.UID(l.ctx)
	if err != nil {
		return nil, err
	}
	requestID, err := parseRequestID(req.RequestId, req.FriendReqId)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.Social.FriendPutInDelete(l.ctx, &social.FriendPutInDeleteReq{
		UserId:      uid,
		ActorUid:    uid,
		FriendReqId: req.FriendReqId,
		RequestId:   requestID,
	})
	if err != nil {
		zLog.Error("friend putIn delete err", zap.Error(err))
		return nil, err
	}
	return
}
