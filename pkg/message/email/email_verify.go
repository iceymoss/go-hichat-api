package mailer

import (
	"sync"
	"time"
)

// VerificationData 验证数据
type VerificationData struct {
	Code      string
	ExpiresAt time.Time
}

// VerificationStore 验证存储
type VerificationStore struct {
	store map[string]VerificationData
	mu    sync.RWMutex
}

// NewVerificationStore 创建新的验证存储
func NewVerificationStore() *VerificationStore {
	return &VerificationStore{
		store: make(map[string]VerificationData),
	}
}

// StoreCode 存储验证码
func (vs *VerificationStore) StoreCode(email, code string, expiration time.Duration) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	vs.store[email] = VerificationData{
		Code:      code,
		ExpiresAt: time.Now().Add(expiration),
	}
}

// VerifyCode 验证验证码
func (vs *VerificationStore) VerifyCode(email, code string) bool {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	data, exists := vs.store[email]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(data.ExpiresAt) {
		return false
	}

	// 检查验证码是否匹配
	return data.Code == code
}

// CleanupExpired 清理过期验证码
func (vs *VerificationStore) CleanupExpired() {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	now := time.Now()
	for email, data := range vs.store {
		if now.After(data.ExpiresAt) {
			delete(vs.store, email)
		}
	}
}
