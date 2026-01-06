package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendMuteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendMuteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendMuteLogic {
	return &FriendMuteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendMuteLogic) FriendMute(in *social.FriendMuteReq) (*social.FriendMuteResp, error) {
	friend, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.FriendUid)
	if err != nil {
		if err == socialmodels.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewReqParamErr(), "friend relation not found uid:%s fid:%s", in.UserId, in.FriendUid)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend relation err:%v req:%v", err, in)
	}

	if in.Muted {
		friend.Muted = 1
	} else {
		friend.Muted = 0
	}

	if err := l.svcCtx.FriendsModel.Update(l.ctx, friend); err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update mute flag err:%v req:%v", err, in)
	}

	return &social.FriendMuteResp{}, nil
}
