package logic

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"
)

// TurnCredential 按 coturn REST(use-auth-secret) 机制签发短期 TURN 凭证。
//
// coturn 用 static-auth-secret 时，服务端不预置用户，而是下发 time-limited 凭证：
//
//	username   = "<过期unix时间戳>"（可选 ":<uid>" 后缀，便于审计）
//	credential = base64( HMAC-SHA1(username, secret) )
//
// coturn 收到后用同一 secret 重算校验，并检查时间戳未过期。凭证短期有效，即使泄露危害有限。
func TurnCredential(secret, uid string, ttl time.Duration, now time.Time) (username, credential string) {
	expiry := now.Add(ttl).Unix()
	username = strconv.FormatInt(expiry, 10)
	if uid != "" {
		username += ":" + uid
	}
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(username))
	credential = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return username, credential
}
