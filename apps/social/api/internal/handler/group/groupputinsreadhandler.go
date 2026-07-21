package group

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/group"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 把我管理的群收到的入群申请全部标记已读
func GroupPutInsReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GroupPutInsReadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := group.NewGroupPutInsReadLogic(r.Context(), svcCtx)
		resp, err := l.GroupPutInsRead(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
