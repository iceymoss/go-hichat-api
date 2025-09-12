package totp_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iceymoss/go-hichat-api/pkg/2fa/totp"
)

// 测试TOTP密钥生成功能
func TestGenerateTOTP(t *testing.T) {
	service := totp.New()

	// 生成TOTP密钥和二维码
	key, qr, err := service.GenerateTOTP("测试应用", "user@example.com")
	require.NoError(t, err, "生成TOTP密钥不应返回错误")

	// 验证返回结果
	assert.NotEmpty(t, key.Secret(), "密钥不应为空")
	assert.NotEmpty(t, key.URL(), "URL不应为空")
	assert.NotEmpty(t, qr, "二维码不应为空")
	assert.Equal(t, "测试应用", key.Issuer(), "发行者名称应匹配")
	assert.Equal(t, "user@example.com", key.AccountName(), "账户名称应匹配")

	fmt.Printf("✅ 密钥: %s\n", key.Secret())

	// 4. 保存为文件
	filename := "totp_qr.png"
	if err = os.WriteFile(filename, qr, 0644); err != nil {
		assert.NoError(t, err, "保存二维码失败")
		return
	}

	// 生成当前验证码
	code, err := service.GenerateCode(key.Secret())
	require.NoError(t, err)

	fmt.Printf("✅ 验证码: %s\n", code)
	// 测试有效验证码
	valid, err := service.ValidateCode(key.Secret(), code)
	require.NoError(t, err)
	assert.True(t, valid, "有效验证码应被接受")
}

func TestGenerateQRCode(t *testing.T) {
	service := totp.New()
	ok, err := service.ValidateCode("SBXIIG3UBCI6YLR5PV6JJPX4V3AVMLSO", "923386")
	assert.NoError(t, err)
	assert.True(t, ok, "验证码有效")
}

// 测试验证码验证功能
func TestValidateCode(t *testing.T) {
	service := totp.New()

	// 生成测试密钥
	key, _, err := service.GenerateTOTP("测试应用", "user@example.com")
	require.NoError(t, err)

	// 生成当前验证码
	code, err := service.GenerateCode(key.Secret())
	require.NoError(t, err)

	fmt.Printf("✅ 验证码: %s\n", code)

	// 测试有效验证码
	valid, err := service.ValidateCode(key.Secret(), code)
	require.NoError(t, err)
	assert.True(t, valid, "有效验证码应被接受")

	// 测试无效验证码
	valid, err = service.ValidateCode(key.Secret(), "234233")
	require.NoError(t, err)
	assert.False(t, valid, "无效验证码应被拒绝")

	// 测试时间偏移补偿
	service.SetTimeOffset(30 * time.Second) // 设置30秒偏移
	nextCode, err := service.GenerateCode(key.Secret())
	require.NoError(t, err)

	// 验证偏移后的验证码
	service.SetTimeOffset(0) // 重置偏移
	valid, err = service.ValidateCode(key.Secret(), nextCode)
	require.NoError(t, err)
	assert.True(t, valid, "应能验证带时间偏移的验证码")
}

// 测试验证码生成功能
func TestGenerateCode(t *testing.T) {
	service := totp.New()

	// 生成密钥
	key, _, err := service.GenerateTOTP("测试应用", "user@example.com")
	require.NoError(t, err)

	// 生成验证码
	code, err := service.GenerateCode(key.Secret())
	require.NoError(t, err)

	// 验证格式
	assert.Len(t, code, 6, "验证码应为6位")

	// 验证生成的验证码
	valid, err := service.ValidateCode(key.Secret(), code)
	require.NoError(t, err)
	assert.True(t, valid, "生成的验证码应有效")
}

// 测试恢复代码生成功能
func TestGenerateRecoveryCodes(t *testing.T) {
	service := totp.New()

	t.Run("有效数量", func(t *testing.T) {
		codes, err := service.GenerateRecoveryCodes(5)
		require.NoError(t, err)
		assert.Len(t, codes, 5, "应生成5个代码")

		// 验证每个代码的格式
		for _, code := range codes {
			assert.Len(t, code, 10, "每个代码应为10个字符")
			assert.Regexp(t, `^[A-Z0-9]{10}$`, code, "代码应为大写字母和数字")
		}
	})

	t.Run("无效数量", func(t *testing.T) {
		_, err := service.GenerateRecoveryCodes(0)
		assert.Error(t, err, "数量小于1应返回错误")

		_, err = service.GenerateRecoveryCodes(21)
		assert.Error(t, err, "数量大于20应返回错误")
	})
}

// 测试恢复代码验证功能
func TestValidateRecoveryCode(t *testing.T) {
	service := totp.New()

	// 准备测试代码
	codes := []string{"ABCDE12345", "FGHIJ67890", "KLMNO24680"}

	t.Run("有效代码", func(t *testing.T) {
		// 验证中间代码
		valid, remaining := service.ValidateRecoveryCode(codes, "FGHIJ67890")
		assert.True(t, valid, "有效代码应被接受")
		assert.ElementsMatch(t, []string{"ABCDE12345", "KLMNO24680"}, remaining, "已用代码应被移除")
	})

	t.Run("无效代码", func(t *testing.T) {
		// 验证不存在的代码
		valid, remaining := service.ValidateRecoveryCode(codes, "INVALID123")
		assert.False(t, valid, "无效代码应被拒绝")
		assert.ElementsMatch(t, codes, remaining, "代码集合应保持不变")
	})

	t.Run("多次使用", func(t *testing.T) {
		remaining := codes
		// 逐个使用所有代码
		for i, code := range codes {
			valid, newRemaining := service.ValidateRecoveryCode(remaining, code)
			assert.True(t, valid, "代码%d应有效", i+1)
			assert.Len(t, newRemaining, len(codes)-i-1, "剩余代码数量应减少")
			remaining = newRemaining
		}

		// 尝试使用已用代码
		valid, remaining := service.ValidateRecoveryCode([]string{}, "ABCDE12345")
		assert.False(t, valid, "已用代码应被拒绝")
		assert.Empty(t, remaining, "应无剩余代码")
	})
}

// 测试安全比较功能（防止时序攻击）
func TestSecureCompare(t *testing.T) {
	service := totp.New()

	t.Run("相同字符串", func(t *testing.T) {
		assert.True(t, service.SecureCompare("测试", "测试"), "相同字符串应匹配")
	})

	t.Run("不同字符串", func(t *testing.T) {
		assert.False(t, service.SecureCompare("测试", "测试1"), "不同字符串不应匹配")
		assert.False(t, service.SecureCompare("测试1", "测试"), "不同字符串不应匹配")
	})

	t.Run("时序攻击防护", func(t *testing.T) {
		// 此测试验证比较时间是否恒定，与输入无关
		short := "短字符串"
		long := "这是一个长得多的字符串用于测试时序攻击防护"

		// 测量比较时间
		measure := func(a, b string) time.Duration {
			start := time.Now()
			for i := 0; i < 100000; i++ {
				service.SecureCompare(a, b)
			}
			return time.Since(start)
		}

		// 比较相同字符串
		timeEqual := measure(short, short)
		// 比较相似字符串
		timeDifferent := measure(short, "短字符X")
		// 比较长度不同的字符串
		timeDifferentLength := measure(short, long)

		// 验证时间差异在10%以内
		assert.InDelta(t, timeEqual, timeDifferent, float64(timeEqual)/10, "比较时间应相似")
		assert.InDelta(t, timeEqual, timeDifferentLength, float64(timeEqual)/10, "比较时间应相似")
	})
}

// 测试剩余时间计算功能
func TestRemainingSeconds(t *testing.T) {
	service := totp.New()

	// 测试不同时间点的剩余时间
	testCases := []struct {
		offset   time.Duration // 时间偏移
		expected int           // 预期剩余秒数
	}{
		{0, 30 - int(time.Now().Unix()%30)},
		{5 * time.Second, 30 - int(time.Now().Add(5*time.Second).Unix()%30)},
		{15 * time.Second, 30 - int(time.Now().Add(15*time.Second).Unix()%30)},
		{25 * time.Second, 30 - int(time.Now().Add(25*time.Second).Unix()%30)},
		{30 * time.Second, 30 - int(time.Now().Add(30*time.Second).Unix()%30)},
	}

	for _, tc := range testCases {
		service.SetTimeOffset(tc.offset)
		remaining := service.RemainingSeconds()

		// 允许1秒误差（测试执行时间）
		assert.True(t, remaining >= tc.expected-1 && remaining <= tc.expected+1,
			"预期约 %d 秒剩余，实际 %d (偏移: %s)", tc.expected, remaining, tc.offset)
	}
}

// 测试时间偏移补偿功能
func TestTimeOffsetCompensation(t *testing.T) {
	service := totp.New()

	// 生成测试密钥
	key, _, err := service.GenerateTOTP("测试应用", "user@example.com")
	require.NoError(t, err)

	// 测试不同时间偏移下的验证
	testCases := []struct {
		offset time.Duration // 时间偏移
		valid  bool          // 预期是否有效
	}{
		{0, true},                  // 无偏移
		{-30 * time.Second, true},  // 前一个周期
		{30 * time.Second, true},   // 后一个周期
		{-31 * time.Second, false}, // 过早（超出范围）
		{31 * time.Second, false},  // 过晚（超出范围）
	}

	for _, tc := range testCases {
		service.SetTimeOffset(tc.offset)

		// 在偏移时间生成验证码
		code, err := service.GenerateCode(key.Secret())
		require.NoError(t, err)

		// 重置时间偏移进行验证
		service.SetTimeOffset(0)

		valid, err := service.ValidateCode(key.Secret(), code)
		require.NoError(t, err)
		assert.Equal(t, tc.valid, valid, "偏移 %s 的验证结果应为 %t", tc.offset, tc.valid)
	}
}
