package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	mailer "github.com/iceymoss/go-hichat-api/pkg/message/email"
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

	cfg := config.ServiceConf.Email
	mailerManager := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	rdb := db.GetRedisConn()
	pass, err := mailerManager.VerifyCode(l.ctx, rdb, in.Email, in.Code)
	if err != nil {
		return nil, libErr.New(xerr.ErrInternalServer, "验证码验证失败")
	}

	if !pass {
		return nil, libErr.New(xerr.ErrInvalidInput, "验证码错误, 请重试")
	}

	return &user.VerifyEmailResponse{}, nil
}
