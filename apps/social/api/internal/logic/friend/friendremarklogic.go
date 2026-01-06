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

type FriendRemarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendRemarkLogic {
	return &FriendRemarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 更新好友备注（仅当前用户视角）
func (l *FriendRemarkLogic) FriendRemark(req *types.FriendRemarkReq) (*types.FriendRemarkResp, error) {
	uid := ctxdata.GetUId(l.ctx)
	if uid == "" {
		return nil, fmt.Errorf("user id not found in context")
	}
	if req.FriendUid == "" {
		return nil, fmt.Errorf("friend_uid is required")
	}

	_, err := l.svcCtx.Social.FriendUpdateRemark(l.ctx, &socialclient.FriendUpdateRemarkReq{
		UserId:    uid,
		FriendUid: req.FriendUid,
		Remark:    req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.FriendRemarkResp{}, nil
}
