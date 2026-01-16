package logic

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendTagsLogic {
	return &FriendTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendTagsLogic) FriendTags(in *social.FriendTagsReq) (*social.FriendTagsResp, error) {
	friend, err := l.svcCtx.FriendsModel.FindByUidAndFid(l.ctx, in.UserId, in.FriendUid)
	if err != nil {
		if err == socialmodels.ErrNotFound {
			return nil, errors.Wrapf(xerr.NewReqParamErr(), "friend relation not found uid:%s fid:%s", in.UserId, in.FriendUid)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friend relation err:%v req:%v", err, in)
	}

	tagBytes, err := json.Marshal(in.Tags)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "marshal tags err:%v req:%v", err, in)
	}

	friend.FriendTags = sql.NullString{
		String: string(tagBytes),
		Valid:  true,
	}
	if err := l.svcCtx.FriendsModel.Update(l.ctx, friend); err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "update friend tags err:%v req:%v", err, in)
	}

	return &social.FriendTagsResp{}, nil
}
