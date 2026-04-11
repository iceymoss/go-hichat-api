package favorite

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/logic/favorite"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// UploadFileHandler 收藏用文件上传
func UploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 限制 50MB
		r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := favorite.NewUploadFileLogic(r.Context(), svcCtx)
		resp, err := l.UploadFile(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
