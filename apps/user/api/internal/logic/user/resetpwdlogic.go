package user

import (
	"context"
	"strings"

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
	// 验证必填字段
	if req.Code == "" || req.Password == "" {
		return nil, errors.New(xerr.ErrBadRequest, "验证码或密码不能为空")
	}

	// 验证至少提供手机号或邮箱之一
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New(xerr.ErrBadRequest, "请提供手机号或邮箱")
	}

	// 验证不能同时提供手机号和邮箱
	if req.Phone != "" && req.Email != "" {
		return nil, errors.New(xerr.ErrBadRequest, "请选择手机号或邮箱其中一种方式，不能同时提供")
	}

	var userInfoResp *user.GetUserByPhoneResponse
	var userInfoRespByEmail *user.GetUserByEmailResponse
	var userId string

	// 根据手机号或邮箱查找用户
	if req.Phone != "" {
		// 使用手机号重置
		userInfoResp, err = l.svcCtx.User.GetUserByPhone(l.ctx, &user.GetUserByPhoneRequest{
			Phone: req.Phone,
		})
		if err != nil {
			// 检查是否是用户不存在的错误
			if err.Error() != "" && (contains(err.Error(), "用户不存在") || contains(err.Error(), "not found")) {
				return nil, errors.New(xerr.ErrNotFound, "该手机号未注册，请先注册账号")
			}
			logx.Errorf("GetUserByPhone failed: %v", err)
			return nil, errors.New(xerr.ErrInternalServer, "查询用户失败，请稍后重试")
		}

		if userInfoResp == nil || userInfoResp.User == nil {
			return nil, errors.New(xerr.ErrNotFound, "该手机号未注册，请先注册账号")
		}

		userId = userInfoResp.User.Id

		// 验证手机验证码
		_, err = l.svcCtx.User.VerifyPhoneCode(l.ctx, &user.VerifyPhoneCodeRequest{
			Phone: req.Phone,
			Code:  req.Code,
		})
		if err != nil {
			// 检查是否是验证码错误
			if contains(err.Error(), "验证码") || contains(err.Error(), "code") {
				return nil, errors.New(xerr.ErrInvalidInput, "手机验证码错误或已过期，请重新获取")
			}
			logx.Errorf("VerifyPhoneCode failed: %v", err)
			return nil, errors.New(xerr.ErrInternalServer, "验证失败，请稍后重试")
		}
	} else {
		// 使用邮箱重置
		userInfoRespByEmail, err = l.svcCtx.User.GetUserByEmail(l.ctx, &user.GetUserByEmailRequest{
			Email: req.Email,
		})
		if err != nil {
			// 检查是否是用户不存在的错误
			if err.Error() != "" && (contains(err.Error(), "用户不存在") || contains(err.Error(), "not found")) {
				return nil, errors.New(xerr.ErrNotFound, "该邮箱未注册，请先注册账号")
			}
			logx.Errorf("GetUserByEmail failed: %v", err)
			return nil, errors.New(xerr.ErrInternalServer, "查询用户失败，请稍后重试")
		}

		if userInfoRespByEmail == nil || userInfoRespByEmail.User == nil {
			return nil, errors.New(xerr.ErrNotFound, "该邮箱未注册，请先注册账号")
		}

		userId = userInfoRespByEmail.User.Id

		// 验证邮件验证码
		_, err = l.svcCtx.User.VerifyEmail(l.ctx, &user.VerifyEmailRequest{
			Email: req.Email,
			Code:  req.Code,
		})
		if err != nil {
			// 检查是否是验证码错误
			if contains(err.Error(), "验证码") || contains(err.Error(), "code") {
				return nil, errors.New(xerr.ErrInvalidInput, "邮箱验证码错误或已过期，请重新获取")
			}
			logx.Errorf("VerifyEmail failed: %v", err)
			return nil, errors.New(xerr.ErrInternalServer, "验证失败，请稍后重试")
		}
	}

	// 更新密码
	_, err = l.svcCtx.User.ResetPassword(l.ctx, &user.ResetPassWordReq{
		Id:       userId,
		Password: req.Password,
	})
	if err != nil {
		logx.Errorf("ResetPassword failed: %v", err)
		return nil, errors.New(xerr.ErrInternalServer, "修改密码失败，请稍后重试")
	}

	return &types.ResetPassWordResp{}, nil
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
