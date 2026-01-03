package verification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	mailer "github.com/iceymoss/go-hichat-api/pkg/message/email"
	"go.uber.org/zap"
)

// EmailCodeSender 邮件验证码发送器适配器
// 将现有的 Mailer 适配为 CodeSender 接口
type EmailCodeSender struct {
	mailer mailer.EmailSender
}

// NewEmailCodeSender 创建邮件验证码发送器
func NewEmailCodeSender(emailSender mailer.EmailSender) *EmailCodeSender {
	return &EmailCodeSender{
		mailer: emailSender,
	}
}

// SendCode 发送验证码（测试阶段打印到控制台）
func (e *EmailCodeSender) SendCode(email, code string) error {
	// 测试阶段：打印到控制台，不真正发送邮件
	logger.Info("发送邮箱验证码", zap.String("email", email), zap.String("code", code))
	fmt.Printf("[Email Code] 邮箱: %s, 验证码: %s\n", email, code)
	
	// TODO: 后续可以集成真实的邮件服务
	// 例如：SendGrid、阿里云邮件推送、腾讯云邮件等
	// return e.mailer.SendVerificationEmail(email, code)
	
	return nil
}

// GenerateCode 生成验证码
func (e *EmailCodeSender) GenerateCode(length int) string {
	return e.mailer.GenerateVerificationCode(length)
}

// SaveCode 保存验证码到Redis
// 直接使用传入的 key（格式：verify:email:xxx@example.com），保持与 CodeSender 接口的一致性
func (e *EmailCodeSender) SaveCode(ctx context.Context, rdb *redis.Client, key, code string) error {
	// 使用SET命令将验证码保存到Redis中，设置5分钟过期时间
	err := rdb.Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}
	return nil
}

// VerifyCode 验证验证码
// 直接使用传入的 key（格式：verify:email:xxx@example.com），保持与 CodeSender 接口的一致性
func (e *EmailCodeSender) VerifyCode(ctx context.Context, rdb *redis.Client, key, code string) (bool, error) {
	// 从Redis获取存储的验证码
	storedCode, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 验证码不存在
			return false, nil
		}
		// 其他错误
		return false, fmt.Errorf("获取验证码失败: %w", err)
	}

	// 比较验证码是否匹配
	if storedCode == code {
		// 验证成功后删除验证码
		rdb.Del(ctx, key)
		return true, nil
	}

	return false, nil
}

// GetCodeType 获取验证码类型
func (e *EmailCodeSender) GetCodeType() CodeType {
	return CodeTypeEmail
}

// extractEmailFromKey 从Redis key中提取email
// key格式：verify:email:xxx@example.com
func extractEmailFromKey(key string) string {
	// 简单实现：假设key格式为 verify:email:xxx
	// 如果key已经是email格式（旧格式），直接返回
	if len(key) > 7 && key[:7] == "verify:" {
		// 新格式：verify:email:xxx
		if len(key) > 13 && key[7:13] == "email:" {
			return key[13:]
		}
		// 旧格式：verify:xxx@example.com
		return key[7:]
	}
	return key
}

