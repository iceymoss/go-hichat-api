package logic

import (
	"context"

	"github.com/iceymoss/go-hichat-api/apps/user/rpc/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/rpc/user"
	pkcCfg "github.com/iceymoss/go-hichat-api/pkg/config"
	"github.com/iceymoss/go-hichat-api/pkg/db"
	libErr "github.com/iceymoss/go-hichat-api/pkg/errors"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/iceymoss/go-hichat-api/pkg/message/verification"
	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type SendPhoneCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendPhoneCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendPhoneCodeLogic {
	return &SendPhoneCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendPhoneCodeLogic) SendPhoneCode(in *user.SendPhoneCodeRequest) (*user.SendPhoneCodeResponse, error) {
	if in.Phone == "" {
		return nil, libErr.New(xerr.ErrBadRequest, "手机号不能为空")
	}

	// 使用工厂模式获取验证码发送器
	codeSender, err := verification.GetCodeSender(verification.CodeTypeSMS)
	if err != nil {
		logger.Error("获取短信验证码发送器失败", zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "获取验证码发送器失败")
	}

	// 生成6位验证码
	code := codeSender.GenerateCode(6)

	// 发送验证码
	err = codeSender.SendCode(in.Phone, code)
	if err != nil {
		logger.Error("发送手机验证码失败", zap.Any("phone", in.Phone), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "发送验证码失败")
	}

	// 保存验证码到Redis
	rdb := db.GetRedisConn()
	key := verification.GetRedisKey(verification.CodeTypeSMS, in.Phone)
	err = codeSender.SaveCode(l.ctx, rdb, key, code)
	if err != nil {
		logger.Error("保存验证码到Redis失败", zap.Any("phone", in.Phone), zap.Error(err))
		return nil, libErr.New(xerr.ErrInternalServer, "保存验证码失败")
	}

	logger.Info("手机验证码已发送", zap.String("phone", in.Phone), zap.String("code", code)) // 打印到控制台

	// 测试/演示模式（未配置真实短信服务商）：把验证码随响应回传，前端可自动填入输入框，
	// 免去真实短信也能完成注册体验；生产环境配置了 aliyun/tencent 等服务商后自动关闭。
	message := "验证码已发送"
	if isDemoSMSMode() {
		message = code
	}

	return &user.SendPhoneCodeResponse{
		Success: true,
		Message: message,
	}, nil
}

// isDemoSMSMode 未配置真实短信服务商（provider 为空或 console）时视为测试/演示模式。
func isDemoSMSMode() bool {
	c := pkcCfg.ServiceConf
	if c == nil || c.Verification == nil {
		return true
	}
	p := c.Verification.SMS.Provider
	return p == "" || p == "console"
}
