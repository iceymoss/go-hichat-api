package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/config"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	mailer "github.com/iceymoss/go-hichat-api/pkg/message/email"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type SendVerificationEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendVerificationEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationEmailLogic {
	return &SendVerificationEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendVerificationEmailLogic) SendVerificationEmail(in *user.SendVerificationEmailRequest) (*user.SendVerificationEmailResponse, error) {
	if in.Email == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "邮箱不能为空")
	}

	cfg := config.ServiceConf.Email
	mailerManager := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	code := mailerManager.GenerateVerificationCode(6)
	err := mailerManager.SendVerificationEmail(in.Email, code)
	if err != nil {
		logger.Error("发送验证码失败", zap.Any("email", in.Email), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "发送验证码失败")
	}

	return &user.SendVerificationEmailResponse{
		Success: true,
		Message: "邮件已发送",
	}, nil
}
