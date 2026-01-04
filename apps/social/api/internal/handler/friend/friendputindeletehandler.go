package friend

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/friend"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除好友申请
func FriendPutInDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FriendPutInDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := friend.NewFriendPutInDeleteLogic(r.Context(), svcCtx)
		resp, err := l.FriendPutInDelete(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
