package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMyGroupMemberSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMyGroupMemberSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMyGroupMemberSettingLogic {
	return &UpdateMyGroupMemberSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMyGroupMemberSettingLogic) UpdateMyGroupMemberSetting(req *types.UpdateMyGroupMemberSettingReq) (resp *types.UpdateMyGroupMemberSettingResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	_, err = l.svcCtx.Social.UpdateMyGroupMemberSetting(l.ctx, &social.UpdateMyGroupMemberSettingReq{
		UserId:        uid,
		GroupId:       req.GroupId,
		GroupNickname: req.GroupNickname,
		GroupRemark:   req.GroupRemark,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateMyGroupMemberSettingResp{}, nil
}
