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

type FriendNotificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendNotificationLogic {
	return &FriendNotificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 消息通知开关
func (l *FriendNotificationLogic) FriendNotification(req *types.FriendNotificationReq) (*types.FriendNotificationResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}

	_, err := l.svcCtx.Social.FriendNotification(l.ctx, &socialclient.FriendNotificationReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		Enabled:   req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendNotificationResp{}, nil
}
