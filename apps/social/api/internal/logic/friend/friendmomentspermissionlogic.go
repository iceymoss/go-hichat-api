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

type FriendMomentsPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendMomentsPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendMomentsPermissionLogic {
	return &FriendMomentsPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 更新朋友圈权限 0允许 1仅聊天 2屏蔽朋友圈
func (l *FriendMomentsPermissionLogic) FriendMomentsPermission(req *types.FriendMomentsPermissionReq) (*types.FriendMomentsPermissionResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}
	if req.Permission < 0 || req.Permission > 2 {
		return nil, fmt.Errorf("permission must be 0/1/2")
	}

	_, err := l.svcCtx.Social.FriendMomentsPermission(l.ctx, &socialclient.FriendMomentsPermissionReq{
		UserId:     uid,
		FriendUid:  req.FriendUid,
		Permission: req.Permission,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendMomentsPermissionResp{}, nil
}
