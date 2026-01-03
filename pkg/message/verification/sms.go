package verification

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
)

// SMSSender 短信验证码发送器
type SMSSender struct {
	// 可以添加短信服务商配置，如阿里云、腾讯云等
	// 目前测试阶段，只打印到控制台
}

// NewSMSSender 创建短信验证码发送器
func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

// SendCode 发送验证码（测试阶段打印到控制台）
func (s *SMSSender) SendCode(phone, code string) error {
	// 测试阶段：打印到控制台
	logger.Info("发送手机验证码", zap.String("phone", phone), zap.String("code", code))
	fmt.Printf("[SMS Code] 手机号: %s, 验证码: %s\n", phone, code)
	
	// TODO: 后续可以集成真实的短信服务商
	// 例如：阿里云短信、腾讯云短信、云片等
	// return s.sendViaAliyun(phone, code)
	
	return nil
}

// GenerateCode 生成6位数字验证码
func (s *SMSSender) GenerateCode(length int) string {
	if length <= 0 {
		length = 6 // 默认6位
	}
	
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

// SaveCode 保存验证码到Redis
func (s *SMSSender) SaveCode(ctx context.Context, rdb *redis.Client, key, code string) error {
	// 使用SET命令将验证码保存到Redis中，设置5分钟过期时间
	err := rdb.Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}
	return nil
}

// VerifyCode 验证验证码是否正确
func (s *SMSSender) VerifyCode(ctx context.Context, rdb *redis.Client, key, code string) (bool, error) {
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
func (s *SMSSender) GetCodeType() CodeType {
	return CodeTypeSMS
}

// GetRedisKey 获取Redis存储键
func GetRedisKey(codeType CodeType, target string) string {
	return fmt.Sprintf("verify:%s:%s", codeType, target)
}

