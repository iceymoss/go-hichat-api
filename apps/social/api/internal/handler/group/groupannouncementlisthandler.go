package group

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/group"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 群公告列表
func GroupAnnouncementListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GroupAnnouncementListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := group.NewGroupAnnouncementListLogic(r.Context(), svcCtx)
		resp, err := l.GroupAnnouncementList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
