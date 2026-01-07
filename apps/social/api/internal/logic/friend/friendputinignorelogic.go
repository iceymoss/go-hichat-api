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

type FriendPutInIgnoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 忽略好友申请
func NewFriendPutInIgnoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInIgnoreLogic {
	return &FriendPutInIgnoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FriendPutInIgnoreLogic) FriendPutInIgnore(req *types.FriendPutInIgnoreReq) (resp *types.FriendPutInIgnoreResp, err error) {
	curUid := l.ctx.Value(Identify).(string)

	// 通过RPC调用忽略好友申请
	_, err = l.svcCtx.Social.FriendPutInIgnore(l.ctx, &social.FriendPutInIgnoreReq{
		FriendReqId: req.FriendReqId,
		UserId:      curUid,
	})
	if err != nil {
		zLog.Error("ignore friend request err", zap.Error(err))
		return nil, err
	}

	return &types.FriendPutInIgnoreResp{}, nil
}
