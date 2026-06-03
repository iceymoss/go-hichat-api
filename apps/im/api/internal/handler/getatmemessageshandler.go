package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取会话中@我且未读的消息列表(用于有人@我快速跳转)
func getAtMeMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAtMeMessagesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetAtMeMessagesLogic(r.Context(), svcCtx)
		resp, err := l.GetAtMeMessages(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
