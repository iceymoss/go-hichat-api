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

type FriendBlockLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendBlockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendBlockLogic {
	return &FriendBlockLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 拉黑/取消拉黑好友（仅当前用户视角）
func (l *FriendBlockLogic) FriendBlock(req *types.FriendBlockReq) (*types.FriendBlockResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}

	_, err := l.svcCtx.Social.FriendBlock(l.ctx, &socialclient.FriendBlockReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		Block:     req.Block,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendBlockResp{}, nil
}
