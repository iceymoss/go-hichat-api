package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 当前用户未读通知数
func getNotificationUnreadCountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetNotificationUnreadCountLogic(r.Context(), svcCtx)
		resp, err := l.GetNotificationUnreadCount()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
