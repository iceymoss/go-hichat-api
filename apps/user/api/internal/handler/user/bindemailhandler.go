package user

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/logic/user"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// BindEmailHandler 绑定/更新邮箱
func BindEmailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindEmailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewBindEmailLogic(r.Context(), svcCtx)
		resp, err := l.BindEmail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

