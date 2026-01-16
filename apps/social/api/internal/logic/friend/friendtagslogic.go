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

type FriendTagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendTagsLogic {
	return &FriendTagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 设置好友标签（覆盖式）
func (l *FriendTagsLogic) FriendTags(req *types.FriendTagsReq) (*types.FriendTagsResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}

	_, err := l.svcCtx.Social.FriendTags(l.ctx, &socialclient.FriendTagsReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		Tags:      req.Tags,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendTagsResp{}, nil
}
