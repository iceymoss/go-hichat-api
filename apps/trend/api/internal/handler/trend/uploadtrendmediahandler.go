package trend

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/logic/trend"
	"github.com/iceymoss/go-hichat-api/apps/trend/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 上传动态媒体
func UploadTrendMediaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxBodySize := int64(500 << 20)
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		if err := r.ParseMultipartForm(maxBodySize); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := trend.NewUploadTrendMediaLogic(r.Context(), svcCtx)
		resp, err := l.UploadTrendMedia(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
