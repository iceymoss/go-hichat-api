package group

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/logic/group"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/svc"
	"github.com/iceymoss/go-hichat-api/apps/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新我的群成员资料（群内昵称/群备注）
func UpdateMyGroupMemberSettingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateMyGroupMemberSettingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := group.NewUpdateMyGroupMemberSettingLogic(r.Context(), svcCtx)
		resp, err := l.UpdateMyGroupMemberSetting(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
