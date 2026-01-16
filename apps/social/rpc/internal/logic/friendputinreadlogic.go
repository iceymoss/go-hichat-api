package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"

	"github.com/zeromicro/go-zero/core/logx"
)

type FriendPutInReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInReadLogic {
	return &FriendPutInReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendPutInReadLogic) FriendPutInRead(in *social.FriendPutInReadReq) (*social.FriendPutInReadResp, error) {
	// 如果 friendReqId 为 0，则标记全部已读
	if in.FriendReqId == 0 {
		err := l.svcCtx.FriendRequestsModel.MarkAllAsRead(l.ctx, in.UserId)
		if err != nil {
			return nil, err
		}
	} else {
		// 标记单条已读
		err := l.svcCtx.FriendRequestsModel.MarkAsRead(l.ctx, in.FriendReqId, in.UserId)
		if err != nil {
			return nil, err
		}
	}

	return &social.FriendPutInReadResp{}, nil
}
