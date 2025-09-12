// Package totp 提供基于时间的一次性密码(TOTP)的完整实现，用于双因素认证(2FA)系统
// 主要功能包括：
//   - TOTP 密钥生成
//   - 二维码生成（用于验证器应用绑定）
//   - TOTP 验证码验证（支持时间偏移补偿）
//   - 恢复代码生成与管理
//   - 安全加密和比较功能
//
// 本包设计用于作为应用程序中实现2FA的核心组件
package totp

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// Service 提供双因素认证的TOTP功能
type Service struct {
	// TimeOffset 用于补偿服务器和客户端设备之间的时钟漂移
	TimeOffset time.Duration
}

// New 创建新的TOTP服务实例
func New() *Service {
	return &Service{}
}

// GenerateTOTP 为用户创建新的TOTP密钥
//
// 参数:
//   - issuer: 应用程序/服务的名称
//   - accountName: 用户标识（邮箱、用户名等）
//
// 返回:
//   - *otp.Key: 生成的TOTP密钥
//   - []byte: QR码图片数据（PNG格式）
//   - error: 生成过程中出现的任何错误
func (s *Service) GenerateTOTP(issuer, accountName string) (*otp.Key, []byte, error) {
	// 使用标准参数生成TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,            // 发行者名称
		AccountName: accountName,       // 账户名称
		SecretSize:  20,                // 160位密钥（20字节）
		Digits:      otp.DigitsSix,     // 6位验证码
		Algorithm:   otp.AlgorithmSHA1, // 使用SHA256算法
	})
	if err != nil {
		return nil, nil, fmt.Errorf("生成TOTP密钥失败: %w", err)
	}

	// 生成绑定用的二维码
	qr, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, nil, fmt.Errorf("生成二维码失败: %w", err)
	}

	return key, qr, nil
}

// ValidateCode 验证TOTP验证码是否有效
//
// 参数:
//   - secret: TOTP密钥
//   - code: 待验证的验证码
//
// 返回:
//   - bool: true表示验证通过，false表示失败
//   - error: 验证过程中出现的任何错误
func (s *Service) ValidateCode(secret, code string) (bool, error) {
	// 应用时间偏移补偿
	//now := time.Now().Add(s.TimeOffset)
	now := time.Now()

	// 使用自定义参数验证验证码
	valid, err := totp.ValidateCustom(
		code,   // 待验证的验证码
		secret, // TOTP密钥
		now,    // 当前时间（带偏移）
		totp.ValidateOpts{
			Period:    30,                // 30秒一个周期
			Skew:      1,                 // 允许前后各1个周期（共3个时间窗口）
			Digits:    otp.DigitsSix,     // 6位验证码
			Algorithm: otp.AlgorithmSHA1, // SHA256算法
		},
	)

	if err != nil {
		return false, fmt.Errorf("验证错误: %w", err)
	}

	return valid, nil
}

// GenerateCode 为给定密钥生成当前时间的TOTP验证码
//
// 参数:
//   - secret: TOTP密钥
//
// 返回:
//   - string: 生成的验证码
//   - error: 生成过程中出现的任何错误
func (s *Service) GenerateCode(secret string) (string, error) {
	// 应用时间偏移补偿
	now := time.Now().Add(s.TimeOffset)

	// 生成当前时间窗口的验证码
	code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    30,                // 30秒周期
		Digits:    otp.DigitsSix,     // 6位验证码
		Algorithm: otp.AlgorithmSHA1, // SHA256算法
	})

	if err != nil {
		return "", fmt.Errorf("生成验证码失败: %w", err)
	}

	return code, nil
}

// GenerateRecoveryCodes 创建一组恢复代码
//
// 参数:
//   - count: 要生成的代码数量
//
// 返回:
//   - []string: 生成的恢复代码
//   - error: 生成过程中出现的任何错误
func (s *Service) GenerateRecoveryCodes(count int) ([]string, error) {
	// 验证数量在合理范围内
	if count < 1 || count > 20 {
		return nil, errors.New("无效的恢复代码数量")
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		// 生成10字节随机数（80位熵）
		randomBytes := make([]byte, 10)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, fmt.Errorf("生成随机字节失败: %w", err)
		}

		// 使用Base32编码并转换为大写
		encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
		codes[i] = strings.ToUpper(encoded)[:10] // 取前10个字符
	}
	return codes, nil
}

// ValidateRecoveryCode 验证恢复代码是否有效，并从集合中移除已使用的代码
//
// 参数:
//   - codes: 当前恢复代码集合
//   - code: 待验证的代码
//
// 返回:
//   - bool: true表示验证通过，false表示失败
//   - []string: 验证后剩余的代码集合
func (s *Service) ValidateRecoveryCode(codes []string, code string) (bool, []string) {
	remainingCodes := []string{} // 存储剩余代码
	found := false               // 标记是否找到匹配代码

	// 遍历所有代码
	for _, c := range codes {
		if !found && s.SecureCompare(c, code) {
			// 找到匹配代码且尚未使用
			found = true
			continue // 跳过已使用的代码
		}
		// 将未使用的代码添加到剩余集合
		remainingCodes = append(remainingCodes, c)
	}

	return found, remainingCodes
}

// SecureCompare 安全比较两个字符串（防止时序攻击）
func (s *Service) SecureCompare(a, b string) bool {
	// 使用恒定时间比较算法
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// RemainingSeconds 计算当前TOTP周期的剩余秒数
func (s *Service) RemainingSeconds() int {
	// 应用时间偏移补偿
	now := time.Now().Add(s.TimeOffset).Unix()
	// 计算当前周期剩余时间
	return 30 - int(now%30)
}

// SetTimeOffset 设置时间偏移补偿值
func (s *Service) SetTimeOffset(offset time.Duration) {
	s.TimeOffset = offset
}
