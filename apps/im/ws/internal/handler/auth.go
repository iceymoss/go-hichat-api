package handler

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/iceymoss/go-hichat-api/apps/im/ws/internal/svc"
	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"
	"github.com/zeromicro/go-zero/rest/token"
	"net/http"
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
	tok, err := auto.parser.ParseToken(r, auto.srvCtx.Config.JwtAuth.AccessSecret, "")
	fmt.Println("token:", tok)
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

	fmt.Println("用户信息:", claims)

	*r = *r.WithContext(context.WithValue(r.Context(), ctxdata.Identify, claims[ctxdata.Identify]))

	return true
}

func (*JwtAuto) UserId(r *http.Request) string {
	return ctxdata.GetUId(r.Context())
}
