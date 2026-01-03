package mailer

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// EmailSender 邮件发送接口
// 通过实现此接口，可以轻松切换不同的邮件发送服务（SMTP、SendGrid、AWS SES等）
type EmailSender interface {
	// SendVerificationEmail 发送验证邮件
	SendVerificationEmail(to, code string) error
	
	// GenerateVerificationCode 生成验证码
	GenerateVerificationCode(length int) string
	
	// SaveVerificationCode 保存验证码到存储（Redis等）
	SaveVerificationCode(ctx context.Context, rdb *redis.Client, email, code string) error
	
	// VerifyCode 验证验证码
	VerifyCode(ctx context.Context, rdb *redis.Client, email, code string) (bool, error)
}

