package group

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyGroupMemberSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyGroupMemberSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyGroupMemberSettingLogic {
	return &GetMyGroupMemberSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyGroupMemberSettingLogic) GetMyGroupMemberSetting(req *types.GetMyGroupMemberSettingReq) (resp *types.GetMyGroupMemberSettingResp, err error) {
	uid := ctxdata.GetUId(l.ctx)
	rpcResp, err := l.svcCtx.Social.GetMyGroupMemberSetting(l.ctx, &social.GetMyGroupMemberSettingReq{
		UserId:  uid,
		GroupId: req.GroupId,
	})
	if err != nil {
		return nil, err
	}

	var setting *types.GroupMemberSetting
	if rpcResp.Setting != nil {
		setting = &types.GroupMemberSetting{
			GroupId:       rpcResp.Setting.GroupId,
			UserId:        rpcResp.Setting.UserId,
			GroupNickname: rpcResp.Setting.GroupNickname,
			GroupRemark:   rpcResp.Setting.GroupRemark,
			UpdatedAt:     rpcResp.Setting.UpdatedAt,
		}
	}

	return &types.GetMyGroupMemberSettingResp{Setting: setting}, nil
}
