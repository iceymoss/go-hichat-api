package user

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPwdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetPwdLogic 重置密码
func NewResetPwdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPwdLogic {
	return &ResetPwdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPwdLogic) ResetPwd(req *types.ResetPassWordReq) (resp *types.ResetPassWordResp, err error) {
	if req.Code == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New(xerr.ErrBadRequest, "邮件验证码或者邮箱或者密码不能为空")
	}

	// 重新用户是否存在
	userInfoResp, err := l.svcCtx.User.GetUserByEmail(l.ctx, &user.GetUserByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		return nil, errors.New(xerr.ErrInternalServer, "查询用户失败")
	}

	if userInfoResp == nil {
		return nil, errors.New(xerr.ErrNotFound, "用户不存在")
	}

	// 验证邮件验证码
	_, err = l.svcCtx.User.VerifyEmail(l.ctx, &user.VerifyEmailRequest{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		return nil, errors.New(xerr.ErrInternalServer, "修改密码失败")
	}

	// 更新密码
	_, err = l.svcCtx.User.ResetPassword(l.ctx, &user.ResetPassWordReq{
		Id:       userInfoResp.User.Id,
		Password: req.Password,
	})
	if err != nil {
		return nil, errors.New(xerr.ErrInternalServer, "修改密码失败")
	}

	return
}
