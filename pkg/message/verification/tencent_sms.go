package verification

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// TencentSMSSender 腾讯云短信发送器
type TencentSMSSender struct {
	secretId   string
	secretKey  string
	sdkAppId   string
	signName   string
	templateId string
}

// NewTencentSMSSender 创建腾讯云短信发送器
func NewTencentSMSSender(config map[string]string) *TencentSMSSender {
	return &TencentSMSSender{
		secretId:   config["secretId"],
		secretKey:  config["secretKey"],
		sdkAppId:   config["sdkAppId"],
		signName:   config["signName"],
		templateId: config["templateId"],
	}
}

// SendCode 发送验证码（通过腾讯云短信服务）
func (t *TencentSMSSender) SendCode(phone, code string) error {
	// TODO: 实现腾讯云短信发送逻辑
	// 这里先使用控制台输出作为占位
	logger.Info("通过腾讯云发送短信验证码", zap.String("phone", phone), zap.String("code", code))
	fmt.Printf("[Tencent SMS] 手机号: %s, 验证码: %s\n", phone, code)
	
	// 实际实现示例：
	// credential := common.NewCredential(t.secretId, t.secretKey)
	// cpf := profile.NewClientProfile()
	// client, _ := sms.NewClient(credential, "ap-guangzhou", cpf)
	// request := sms.NewSendSmsRequest()
	// request.PhoneNumberSet = []*string{&phone}
	// request.TemplateID = &t.templateId
	// request.Sign = &t.signName
	// request.TemplateParamSet = []*string{&code}
	// request.SmsSdkAppid = &t.sdkAppId
	// _, err := client.SendSms(request)
	// return err
	
	return nil
}

// GenerateCode 生成验证码
func (t *TencentSMSSender) GenerateCode(length int) string {
	return NewSMSSender().GenerateCode(length)
}

// SaveCode 保存验证码到Redis
func (t *TencentSMSSender) SaveCode(ctx context.Context, rdb *redis.Client, key, code string) error {
	return NewSMSSender().SaveCode(ctx, rdb, key, code)
}

// VerifyCode 验证验证码
func (t *TencentSMSSender) VerifyCode(ctx context.Context, rdb *redis.Client, key, code string) (bool, error) {
	return NewSMSSender().VerifyCode(ctx, rdb, key, code)
}

// GetCodeType 获取验证码类型
func (t *TencentSMSSender) GetCodeType() CodeType {
	return CodeTypeSMS
}

