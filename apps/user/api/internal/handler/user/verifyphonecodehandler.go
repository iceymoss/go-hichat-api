package user

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/logic/user"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 验证手机验证码
func VerifyPhoneCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VerifyPhoneCodeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewVerifyPhoneCodeLogic(r.Context(), svcCtx)
		resp, err := l.VerifyPhoneCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

