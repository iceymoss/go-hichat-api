package verification

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// CodeType 验证码类型
type CodeType string

const (
	CodeTypeSMS   CodeType = "sms"   // 手机验证码
	CodeTypeEmail CodeType = "email" // 邮箱验证码
)

// CodeSender 验证码发送接口
// 通过实现此接口，可以轻松切换不同的验证码发送服务（短信、邮件等）
type CodeSender interface {
	// SendCode 发送验证码
	// to: 接收方（手机号或邮箱）
	// code: 验证码
	SendCode(to, code string) error

	// GenerateCode 生成验证码
	// length: 验证码长度（默认6位）
	GenerateCode(length int) string

	// SaveCode 保存验证码到存储（Redis等）
	// key: 存储键（通常是 "verify:phone:xxx" 或 "verify:email:xxx"）
	// code: 验证码
	SaveCode(ctx context.Context, rdb *redis.Client, key, code string) error

	// VerifyCode 验证验证码
	// key: 存储键
	// code: 待验证的验证码
	VerifyCode(ctx context.Context, rdb *redis.Client, key, code string) (bool, error)

	// GetCodeType 获取验证码类型
	GetCodeType() CodeType
}
