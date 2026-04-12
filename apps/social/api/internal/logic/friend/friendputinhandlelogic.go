package friend

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
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
	curUid := l.ctx.Value(Identify).(string)
	_, err = l.svcCtx.Social.FriendPutInHandle(l.ctx, &social.FriendPutInHandleReq{
		FriendReqId:  req.FriendReqId,
		UserId:       curUid,
		HandleResult: req.HandleResult,
		Remark:       req.Remark,
		HandleMsg:    req.HandleMsg,
		Tags:         req.Tags,
	})
	if err != nil {
		zLog.Error("friend req handle err", zap.Error(err))
		return nil, err
	}

	return
}
