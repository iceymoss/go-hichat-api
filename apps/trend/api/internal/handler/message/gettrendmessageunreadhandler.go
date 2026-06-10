package message

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/logic/message"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取动态消息未读数（总数+按类型明细）
func GetTrendMessageUnreadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := message.NewGetTrendMessageUnreadLogic(r.Context(), svcCtx)
		resp, err := l.GetTrendMessageUnread()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
