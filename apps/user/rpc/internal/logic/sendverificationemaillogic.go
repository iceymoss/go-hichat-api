package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/message/verification"
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

	// 使用工厂模式获取验证码发送器
	codeSender, err := verification.GetCodeSender(verification.CodeTypeEmail)
	if err != nil {
		logger.Error("获取邮件验证码发送器失败", zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "获取验证码发送器失败")
	}
	
	// 生成验证码
	code := codeSender.GenerateCode(6)
	
	// 保存验证码到Redis
	rdb := db.GetRedisConn()
	key := verification.GetRedisKey(verification.CodeTypeEmail, in.Email)
	err = codeSender.SaveCode(l.ctx, rdb, key, code)
	if err != nil {
		logger.Error("保存验证码失败", zap.Any("email", in.Email), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "保存验证码失败")
	}
	
	// 发送邮件
	err = codeSender.SendCode(in.Email, code)
	if err != nil {
		logger.Error("发送验证码失败", zap.Any("email", in.Email), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "发送验证码失败")
	}
	
	// 测试环境：在控制台打印验证码
	logger.Info("邮件验证码（测试用）", zap.String("email", in.Email), zap.String("code", code))
	logx.Infof("邮件验证码（测试用） - Email: %s, Code: %s", in.Email, code)

	return &user.SendVerificationEmailResponse{
		Success: true,
		Message: "邮件已发送",
	}, nil
}
