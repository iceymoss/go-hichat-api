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

type FriendUpdateRemarkLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendUpdateRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendUpdateRemarkLogic {
	return &FriendUpdateRemarkLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendUpdateRemarkLogic) FriendUpdateRemark(in *social.FriendUpdateRemarkReq) (*social.FriendUpdateRemarkResp, error) {
	friend, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.FriendUid)
	if err != nil {
		if err == socialmodels.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewReqParamErr(), "friend relation not found uid:%s fid:%s", in.UserId, in.FriendUid)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend relation err:%v req:%v", err, in)
	}

	friend.Remark = in.Remark
	if err := l.svcCtx.FriendsModel.Update(l.ctx, friend); err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update friend remark err:%v req:%v", err, in)
	}

	return &social.FriendUpdateRemarkResp{}, nil
}
