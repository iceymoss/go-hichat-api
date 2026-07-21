package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GroupCallStateResp /v1/streaming/group-call 响应：某群当前是否有活跃通话 + 参与者。
// 用于进群聊补拉（错过 group.state 广播时）/ 加入通话按钮。
type GroupCallStateResp struct {
	Active       bool     `json:"active"`
	CallId       string   `json:"callId,omitempty"`
	CallType     string   `json:"callType,omitempty"`
	Participants []string `json:"participants"`
}

// GroupCallStateHandler 查询某群当前活跃通话状态（需鉴权 + 群成员）。
func (s *SignalingServer) GroupCallStateHandler(auth *JwtAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		uid := auth.ParseUID(r)
		if uid == "" {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		groupID := r.URL.Query().Get("group_id")
		if groupID == "" {
			httpx.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "group_id required"})
			return
		}
		if !s.isGroupMember(uid, groupID) {
			httpx.WriteJson(w, http.StatusForbidden, map[string]string{"error": "not a group member"})
			return
		}

		callID, t, participants, ok := s.groups.ActiveByGroup(groupID)
		resp := GroupCallStateResp{Active: ok, Participants: participants}
		if ok {
			resp.CallId = callID
			resp.CallType = string(t)
		} else {
			resp.Participants = []string{}
		}
		httpx.WriteJson(w, http.StatusOK, resp)
	}
}
