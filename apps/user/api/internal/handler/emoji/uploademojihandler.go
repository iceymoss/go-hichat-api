package emoji

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/logic/emoji"
	"github.com/iceymoss/go-hichat-api/apps/user/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func UploadEmojiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := emoji.NewUploadEmojiLogic(r.Context(), svcCtx)
		resp, err := l.UploadEmoji(file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
