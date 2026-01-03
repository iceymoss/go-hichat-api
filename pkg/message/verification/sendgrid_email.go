package verification

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/pkg/config"
	mailer "github.com/iceymoss/go-hichat-api/pkg/message/email"
)

// SendGridEmailSender SendGrid邮件发送器
type SendGridEmailSender struct {
	apiKey string
	from   string
}

// NewSendGridEmailSender 创建SendGrid邮件发送器
func NewSendGridEmailSender(config map[string]string) *SendGridEmailSender {
	return &SendGridEmailSender{
		apiKey: config["apiKey"],
		from:   config["from"],
	}
}

// SendCode 发送验证码（通过SendGrid）
func (s *SendGridEmailSender) SendCode(email, code string) error {
	// TODO: 实现SendGrid邮件发送逻辑
	// 实际实现示例：
	// message := mail.NewSingleEmail(
	//     mail.NewEmail("HiChat", s.from),
	//     "您的验证码",
	//     mail.NewEmail("", email),
	//     fmt.Sprintf("您的验证码是: %s", code),
	//     fmt.Sprintf("<html><body><h2>您的验证码是: <strong>%s</strong></h2></body></html>", code),
	// )
	// client := sendgrid.NewSendClient(s.apiKey)
	// _, err := client.Send(message)
	// return err
	
	// 临时使用SMTP作为后备方案
	cfg := config.ServiceConf.Email
	emailSender := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return emailSender.SendVerificationEmail(email, code)
}

// GenerateCode 生成验证码
func (s *SendGridEmailSender) GenerateCode(length int) string {
	cfg := config.ServiceConf.Email
	emailSender := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return emailSender.GenerateVerificationCode(length)
}

// SaveCode 保存验证码到Redis
func (s *SendGridEmailSender) SaveCode(ctx context.Context, rdb *redis.Client, key, code string) error {
	email := extractEmailFromKey(key)
	cfg := config.ServiceConf.Email
	emailSender := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return emailSender.SaveVerificationCode(ctx, rdb, email, code)
}

// VerifyCode 验证验证码
func (s *SendGridEmailSender) VerifyCode(ctx context.Context, rdb *redis.Client, key, code string) (bool, error) {
	email := extractEmailFromKey(key)
	cfg := config.ServiceConf.Email
	emailSender := mailer.NewMailer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return emailSender.VerifyCode(ctx, rdb, email, code)
}

// GetCodeType 获取验证码类型
func (s *SendGridEmailSender) GetCodeType() CodeType {
	return CodeTypeEmail
}

