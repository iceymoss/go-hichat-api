package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/social/socialmodels"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/rpc/social"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMemberRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMemberRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMemberRoleLogic {
	return &GetMemberRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetMemberRole 点查某用户在某群的角色等级。
// 不在群（ErrNotFound）不是错误，返回 isMember=false；其余 DB 错误才上抛。
func (l *GetMemberRoleLogic) GetMemberRole(in *social.GetMemberRoleReq) (*social.GetMemberRoleResp, error) {
	member, err := l.svcCtx.GroupMembersModel.FindByGroudIdAndUserId(l.ctx, in.UserId, in.GroupId)
	if err != nil {
		if errors.Is(err, socialmodels.ErrNotFound) {
			return &social.GetMemberRoleResp{IsMember: false}, nil
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "get member role err: %v, groupId: %s, userId: %s", err, in.GroupId, in.UserId)
	}
	return &social.GetMemberRoleResp{
		RoleLevel: int32(member.RoleLevel),
		IsMember:  true,
	}, nil
}
