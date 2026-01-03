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

type VerifyPhoneCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyPhoneCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyPhoneCodeLogic {
	return &VerifyPhoneCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyPhoneCodeLogic) VerifyPhoneCode(in *user.VerifyPhoneCodeRequest) (*user.VerifyPhoneCodeResponse, error) {
	if in.Phone == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "手机号不能为空")
	}

	if in.Code == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "验证码不能为空")
	}

	// 使用工厂模式获取验证码发送器
	codeSender, err := verification.GetCodeSender(verification.CodeTypeSMS)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "获取验证码发送器失败")
	}
	
	rdb := db.GetRedisConn()
	key := verification.GetRedisKey(verification.CodeTypeSMS, in.Phone)
	
	pass, err := codeSender.VerifyCode(l.ctx, rdb, key, in.Code)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "验证码验证失败")
	}

	if !pass {
		return nil, libErr.New(xerr.ErrInvalidInput, "验证码错误或已过期")
	}

	return &user.VerifyPhoneCodeResponse{
		Success: true,
		Message: "验证成功",
	}, nil
}

