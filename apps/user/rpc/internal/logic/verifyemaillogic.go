package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/message/verification"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyEmailLogic {
	return &VerifyEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyEmailLogic) VerifyEmail(in *user.VerifyEmailRequest) (*user.VerifyEmailResponse, error) {
	if in.Email == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "邮箱不能为空")
	}

	if in.Code == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "验证码不能为空")
	}

	// 使用工厂模式获取验证码发送器
	codeSender, err := verification.GetCodeSender(verification.CodeTypeEmail)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "获取验证码发送器失败")
	}
	
	rdb := db.GetRedisConn()
	key := verification.GetRedisKey(verification.CodeTypeEmail, in.Email)
	pass, err := codeSender.VerifyCode(l.ctx, rdb, key, in.Code)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "验证码验证失败")
	}

	if !pass {
		return nil, libErr.New(xerr.ErrInvalidInput, "验证码错误, 请重试")
	}

	return &user.VerifyEmailResponse{
		Success: true,
	}, nil
}
