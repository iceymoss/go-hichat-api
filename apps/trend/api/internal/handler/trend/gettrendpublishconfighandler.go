package trend

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/logic/trend"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取动态发布配置
func GetTrendPublishConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := trend.NewGetTrendPublishConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetTrendPublishConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
