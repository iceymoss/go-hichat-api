package trend

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/logic/trend"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetLatestTrendsHandler 获取最新动态流
func GetLatestTrendsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetLatestTrendsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := trend.NewGetLatestTrendsLogic(r.Context(), svcCtx)
		resp, err := l.GetLatestTrends(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
