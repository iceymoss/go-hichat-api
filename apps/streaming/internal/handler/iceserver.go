package handler

import (
	"net/http"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ICEServer 单个 ICE 服务器（与浏览器 RTCIceServer 对齐）
type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// ICEServersResp /v1/streaming/ice-servers 响应（前端直接喂给 RTCPeerConnection）
type ICEServersResp struct {
	IceServers []ICEServer `json:"iceServers"`
}

// ICEServersHandler 下发 ICE（STUN/TURN）配置给前端。
// 本期返回静态 STUN；接入 TURN 后此处下发短期 TURN 凭证（HMAC time-limited）。
func ICEServersHandler(svcCtx *svc.ServiceContext, auth *JwtAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if uid := auth.ParseUID(r); uid == "" {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		resp := ICEServersResp{IceServers: make([]ICEServer, 0, len(svcCtx.Config.WebRTC.IceServers))}
		for _, s := range svcCtx.Config.WebRTC.IceServers {
			resp.IceServers = append(resp.IceServers, ICEServer{
				URLs:       s.URLs,
				Username:   s.Username,
				Credential: s.Credential,
			})
		}

		httpx.OkJson(w, resp)
	}
}
