package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendShareLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendShareLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendShareLogic {
	return &FriendShareLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendShareLogic) FriendShare(in *social.FriendShareReq) (*social.FriendShareResp, error) {
	if in.UserId == "" || in.FriendUid == "" {
		return nil, errors.Wrapf(xerr.NewReqParamErr(), "missing share params req:%v", in)
	}
	// 预留：当前仅返回成功，后续可接入分享记录或推送

	return &social.FriendShareResp{}, nil
}
