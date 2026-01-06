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

type FriendMomentsPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendMomentsPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendMomentsPermissionLogic {
	return &FriendMomentsPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendMomentsPermissionLogic) FriendMomentsPermission(in *social.FriendMomentsPermissionReq) (*social.FriendMomentsPermissionResp, error) {
	if in.Permission < 0 || in.Permission > 2 {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "invalid permission:%d req:%v", in.Permission, in)
	}

	friend, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.FriendUid)
	if err != nil {
		if err == socialmodels.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewReqParamErr(), "friend relation not found uid:%s fid:%s", in.UserId, in.FriendUid)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend relation err:%v req:%v", err, in)
	}

	friend.MomentsPermission = int(in.Permission)
	if err := l.svcCtx.FriendsModel.Update(l.ctx, friend); err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update moments permission err:%v req:%v", err, in)
	}

	return &social.FriendMomentsPermissionResp{}, nil
}
