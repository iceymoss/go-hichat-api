package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInDeleteLogic {
	return &FriendPutInDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// FriendPutInDelete 软删除好友申请记录（将 status 设为 0）
func (l *FriendPutInDeleteLogic) FriendPutInDelete(in *social.FriendPutInDeleteReq) (*social.FriendPutInDeleteResp, error) {
	if in.FriendReqId <= 0 {
		return nil, errors.Wrapf(xerr.NewMsg("申请记录ID不能为空"), "friend_req_id is required")
	}

	err := l.svcCtx.FriendRequestsModel.SoftDelete(l.ctx, uint64(in.FriendReqId), in.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "soft delete friend request err %v, reqId %v", err, in.FriendReqId)
	}

	return &social.FriendPutInDeleteResp{}, nil
}
