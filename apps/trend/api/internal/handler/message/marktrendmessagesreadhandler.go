package message

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/logic/message"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 动态消息全部标记为已读
func MarkTrendMessagesReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := message.NewMarkTrendMessagesReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkTrendMessagesRead()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
