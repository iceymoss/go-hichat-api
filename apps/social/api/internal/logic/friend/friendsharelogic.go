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

type FriendShareLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendShareLogic {
	return &FriendShareLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 分享好友（预留：当前仅返回成功，不做存储）
func (l *FriendShareLogic) FriendShare(req *types.FriendShareReq) (*types.FriendShareResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}
	_, err := l.svcCtx.Social.FriendShare(l.ctx, &socialclient.FriendShareReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		TargetUid: req.TargetUid,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendShareResp{}, nil
}
