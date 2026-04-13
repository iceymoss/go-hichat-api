package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyGroupMemberSettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyGroupMemberSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyGroupMemberSettingLogic {
	return &GetMyGroupMemberSettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetMyGroupMemberSetting 获取我的群设置（群昵称/群备注），直接从 group_members 表读取
func (l *GetMyGroupMemberSettingLogic) GetMyGroupMemberSetting(in *social.GetMyGroupMemberSettingReq) (*social.GetMyGroupMemberSettingResp, error) {
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewMsg("不在群中"), "not in group")
	}

	return &social.GetMyGroupMemberSettingResp{
		Setting: &social.GroupMemberSetting{
			GroupId:       in.GroupId,
			UserId:        in.UserId,
			GroupNickname: member.GroupNickname,
			GroupRemark:   member.GroupRemark,
		},
	}, nil
}
