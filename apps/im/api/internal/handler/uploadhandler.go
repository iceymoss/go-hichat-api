package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/logic"
	"github.com/iceymoss/go-hichat-api/apps/im/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 上传富媒体文件(图片/视频/文件/语音)
func uploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 multipart 表单，上限 100MB
		if err := r.ParseMultipartForm(100 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := logic.NewUploadLogic(r.Context(), svcCtx)
		resp, err := l.Upload(file, fileHeader)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
