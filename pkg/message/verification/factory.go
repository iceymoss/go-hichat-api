package verification

import (
	"fmt"

	"github.com/iceymoss/go-hichat-api/pkg/config"
	mailer "github.com/iceymoss/go-hichat-api/pkg/message/email"
)

// CodeSenderFactory 验证码发送器工厂
// 根据配置创建对应的验证码发送器实现
type CodeSenderFactory struct {
	smsConfig   SMSConfig
	emailConfig EmailConfig
}

// SMSConfig 短信服务配置
type SMSConfig struct {
	Provider string            `mapstructure:"provider" json:"provider"` // 服务商：console/aliyun/tencent
	Config   map[string]string  `mapstructure:"config" json:"config"`    // 服务商特定配置
}

// EmailConfig 邮件服务配置
type EmailConfig struct {
	Provider string            `mapstructure:"provider" json:"provider"` // 服务商：smtp/sendgrid
	Config   map[string]string `mapstructure:"config" json:"config"`     // 服务商特定配置
}

// NewCodeSenderFactory 创建验证码发送器工厂
func NewCodeSenderFactory() *CodeSenderFactory {
	factory := &CodeSenderFactory{}
	
	// 从配置中读取验证码服务配置
	if config.ServiceConf != nil {
		if cfg := config.ServiceConf.Verification; cfg != nil {
			factory.smsConfig = SMSConfig{
				Provider: cfg.SMS.Provider,
				Config:   cfg.SMS.Config,
			}
			factory.emailConfig = EmailConfig{
				Provider: cfg.Email.Provider,
				Config:   cfg.Email.Config,
			}
		}
	}
	
	return factory
}

// GetCodeSender 根据验证码类型获取对应的发送器
// codeType: 验证码类型（sms/email）
func (f *CodeSenderFactory) GetCodeSender(codeType CodeType) (CodeSender, error) {
	switch codeType {
	case CodeTypeSMS:
		return f.createSMSSender()
	case CodeTypeEmail:
		return f.createEmailSender()
	default:
		return nil, fmt.Errorf("不支持的验证码类型: %s", codeType)
	}
}

// createSMSSender 创建短信验证码发送器
func (f *CodeSenderFactory) createSMSSender() (CodeSender, error) {
	provider := f.smsConfig.Provider
	if provider == "" {
		provider = "console" // 默认使用控制台输出（测试用）
	}
	
	switch provider {
	case "console":
		// 控制台输出（测试用）
		return NewSMSSender(), nil
	case "aliyun":
		// 阿里云短信服务
		return NewAliyunSMSSender(f.smsConfig.Config), nil
	case "tencent":
		// 腾讯云短信服务
		return NewTencentSMSSender(f.smsConfig.Config), nil
	default:
		return nil, fmt.Errorf("不支持的短信服务商: %s", provider)
	}
}

// createEmailSender 创建邮件验证码发送器
func (f *CodeSenderFactory) createEmailSender() (CodeSender, error) {
	provider := f.emailConfig.Provider
	if provider == "" {
		provider = "smtp" // 默认使用SMTP
	}
	
	switch provider {
	case "smtp":
		// SMTP邮件服务
		cfg := config.ServiceConf.Email
		emailSender := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
		return NewEmailCodeSender(emailSender), nil
	case "sendgrid":
		// SendGrid邮件服务
		return NewSendGridEmailSender(f.emailConfig.Config), nil
	default:
		return nil, fmt.Errorf("不支持的邮件服务商: %s", provider)
	}
}

// 全局工厂实例（懒加载）
var globalFactory *CodeSenderFactory

// GetCodeSender 全局函数：获取验证码发送器（推荐使用）
// 内部使用全局工厂实例，首次调用时初始化
func GetCodeSender(codeType CodeType) (CodeSender, error) {
	if globalFactory == nil {
		globalFactory = NewCodeSenderFactory()
	}
	return globalFactory.GetCodeSender(codeType)
}

