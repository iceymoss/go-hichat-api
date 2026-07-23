package handler

import (
	"context"
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/rest/token"
)

type JwtAuto struct {
	srvCtx *svc.ServiceContext
	parser *token.TokenParser
}

func NewJwtAuto(srvCtx *svc.ServiceContext) *JwtAuto {
	return &JwtAuto{
		srvCtx: srvCtx,
		parser: token.NewTokenParser(),
	}
}

func (auto *JwtAuto) Auth(w http.ResponseWriter, r *http.Request) bool {
	// 浏览器原生 WebSocket 不支持自定义 Header，
	// 需要支持从 URL 查询参数 ?token=xxx 读取 JWT
	if qToken := r.URL.Query().Get("token"); qToken != "" && r.Header.Get("Authorization") == "" {
		r.Header.Set("Authorization", "Bearer "+qToken)
	}

	tok, err := auto.parser.ParseToken(r, auto.srvCtx.Config.JwtAuth.AccessSecret, "")
	if err != nil {
		return false
	}
	if !tok.Valid {
		return false
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	*r = *r.WithContext(context.WithValue(r.Context(), ctxdata.Identify, claims[ctxdata.Identify]))

	return true
}

func (*JwtAuto) UserId(r *http.Request) string {
	return ctxdata.GetUId(r.Context())
}
