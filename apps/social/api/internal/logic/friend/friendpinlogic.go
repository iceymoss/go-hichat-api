package friend

import (
	"context"
	"fmt"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/socialclient"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPinLogic {
	return &FriendPinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 置顶开关
func (l *FriendPinLogic) FriendPin(req *types.FriendPinReq) (*types.FriendPinResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}

	_, err := l.svcCtx.Social.FriendPin(l.ctx, &socialclient.FriendPinReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		Pinned:    req.Pinned,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendPinResp{}, nil
}
