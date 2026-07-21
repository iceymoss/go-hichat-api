package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 撤回消息
func recallMsgHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RecallMsgReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewRecallMsgLogic(r.Context(), svcCtx)
		resp, err := l.RecallMsg(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
