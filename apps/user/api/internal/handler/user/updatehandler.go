package user

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/logic/user"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 修改用户信息
func UpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := user.NewUpdateLogic(r.Context(), svcCtx)
		resp, err := l.Update(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
