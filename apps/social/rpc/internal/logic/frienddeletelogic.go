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

type FriendDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendDeleteLogic {
	return &FriendDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendDeleteLogic) FriendDelete(in *social.FriendDeleteReq) (*social.FriendDeleteResp, error) {
	var idsToDelete []uint64

	relation, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.FriendUid)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend relation err:%v req:%v", err, in)
	}
	if relation != nil {
		idsToDelete = append(idsToDelete, relation.Id)
	}

	reverse, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.FriendUid, in.UserId)
	if err != nil && err != socialmodels.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find reverse friend relation err:%v req:%v", err, in)
	}
	if reverse != nil {
		idsToDelete = append(idsToDelete, reverse.Id)
	}

	if len(idsToDelete) == 0 {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "friend relation not found uid:%s fid:%s", in.UserId, in.FriendUid)
	}

	for _, id := range idsToDelete {
		if err := l.svcCtx.FriendsModel.Delete(l.ctx, id); err != nil {
			return nil, errors.Wrapf(xerr.NewDBErr(), "delete friend relation err:%v id:%d req:%v", err, id, in)
		}
	}

	return &social.FriendDeleteResp{}, nil
}
