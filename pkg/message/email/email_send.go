package mailer

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/go-redis/redis/v8"
)

// Mailer 邮件发送器（实现 EmailSender 接口）
type Mailer struct {
	Host     string
	Port     string
	Username string
	Password string
}

// 确保 Mailer 实现了 EmailSender 接口
var _ EmailSender = (*Mailer)(nil)

// NewMailer 创建邮件发送器
func NewMailer(host, port, username, password string) *Mailer {
	return &Mailer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

// SendVerificationEmail 发送验证邮件
func (m *Mailer) SendVerificationEmail(to, code string) error {
	subject := "您的验证码"
	body := fmt.Sprintf(`
	<html>
	<body>
		<h2>HiChat 邮箱验证</h2>
		<p>您的验证码是: <strong>%s</strong></p>
		<p>请在5分钟内使用此验证码完成验证。</p>
		<p>如果不是您本人操作，请忽略此邮件。</p>
	</body>
	</html>
	`, code)

	return m.sendEmail(to, subject, body)
}

// sendEmail 实际发送邮件
func (m *Mailer) sendEmail(to, subject, body string) error {
	// 邮件头部
	headers := make(map[string]string)
	headers["From"] = m.Username
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// 身份验证
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	// 发送邮件（带重试）
	const maxRetries = 2
	var err error

	for i := 0; i <= maxRetries; i++ {
		err = smtp.SendMail(
			m.Host+":"+m.Port,
			auth,
			m.Username,
			[]string{to},
			[]byte(message),
		)

		if err == nil {
			return nil
		}

		if i < maxRetries {
			log.Printf("邮件发送失败，重试中 (%d/%d): %v", i+1, maxRetries, err)
			time.Sleep(1 * time.Second) // 等待后重试
		}
	}

	return fmt.Errorf("邮件发送失败: %w", err)
}

// GenerateVerificationCode 生成随机验证码
func (m *Mailer) GenerateVerificationCode(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// 如果随机源失败，使用时间戳作为后备方案
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// SaveVerificationCode 将验证码保存到Redis
func (m *Mailer) SaveVerificationCode(ctx context.Context, rdb *redis.Client, email, code string) error {
	// 使用SET命令将验证码保存到Redis中，设置5分钟过期时间
	err := rdb.Set(ctx, "verify:"+email, code, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	return nil
}

// VerifyCode 验证验证码是否正确
func (m *Mailer) VerifyCode(ctx context.Context, rdb *redis.Client, email, code string) (bool, error) {
	// 从Redis获取存储的验证码
	storedCode, err := rdb.Get(ctx, "verify:"+email).Result()
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
		rdb.Del(ctx, "verify:"+email)
		return true, nil
	}

	return false, nil
}
