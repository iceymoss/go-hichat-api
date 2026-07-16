package handler

import (
	"net/http"
	"time"

	"github.com/iceymoss/go-hichat-api/apps/streaming/internal/logic"
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

// defaultSTUN 兜底 STUN：go-zero 配置加载对 yaml tag 不稳，且同机跨浏览器需要 srflx
// 候选绕开 mDNS（.local）host 候选解析失败的问题。配置为空时至少给公共 STUN。
var defaultSTUN = []ICEServer{
	{URLs: []string{"stun:stun.l.google.com:19302"}},
	{URLs: []string{"stun:stun1.l.google.com:19302"}},
}

// ICEServersHandler 下发 ICE（STUN/TURN）配置给前端。
// 本期返回静态 STUN；接入 TURN 后此处下发短期 TURN 凭证（HMAC time-limited）。
func ICEServersHandler(svcCtx *svc.ServiceContext, auth *JwtAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 跨端口（web :3001 -> streaming :10093）需要 CORS。token 走 query，属简单请求；
		// 仍处理 OPTIONS 预检以兼容带头部的情况。
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

		resp := ICEServersResp{IceServers: make([]ICEServer, 0, len(svcCtx.Config.WebRTC.IceServers)+1)}
		for _, s := range svcCtx.Config.WebRTC.IceServers {
			resp.IceServers = append(resp.IceServers, ICEServer{
				URLs:       s.URLs,
				Username:   s.Username,
				Credential: s.Credential,
			})
		}
		// coturn TURN：为本次请求签发短期凭证（HMAC，避免下发长期密钥）
		if turn := svcCtx.Config.Turn; turn.Secret != "" && len(turn.URLs) > 0 {
			ttl := time.Duration(turn.TTLSeconds) * time.Second
			if ttl <= 0 {
				ttl = time.Hour
			}
			user, cred := logic.TurnCredential(turn.Secret, uid, ttl, time.Now())
			resp.IceServers = append(resp.IceServers, ICEServer{
				URLs:       turn.URLs,
				Username:   user,
				Credential: cred,
			})
		}
		if len(resp.IceServers) == 0 {
			resp.IceServers = defaultSTUN
		}

		httpx.OkJson(w, resp)
	}
}

