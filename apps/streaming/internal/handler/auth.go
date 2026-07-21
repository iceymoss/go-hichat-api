package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/rest/token"
)

// JwtAuth 校验 streaming 服务（ws 升级 + HTTP 接口）的 JWT，
// 与 im/user 共用同一 AccessSecret，照搬 apps/im/ws 的解析方式。
type JwtAuth struct {
	svcCtx *svc.ServiceContext
	parser *token.TokenParser
}

func NewJwtAuth(svcCtx *svc.ServiceContext) *JwtAuth {
	return &JwtAuth{
		svcCtx: svcCtx,
		parser: token.NewTokenParser(),
	}
}

// ParseUID 解析请求中的 JWT，返回用户 ID。
// 浏览器原生 WebSocket 不支持自定义 Header，支持从 ?token=xxx 读取。
// 返回空串表示鉴权失败。
func (a *JwtAuth) ParseUID(r *http.Request) string {
	if qToken := r.URL.Query().Get("token"); qToken != "" && r.Header.Get("Authorization") == "" {
		r.Header.Set("Authorization", "Bearer "+qToken)
	}

	tok, err := a.parser.ParseToken(r, a.svcCtx.Config.JwtAuth.AccessSecret, "")
	if err != nil || !tok.Valid {
		return ""
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	uid, _ := claims[ctxdata.Identify].(string)
	return uid
}
