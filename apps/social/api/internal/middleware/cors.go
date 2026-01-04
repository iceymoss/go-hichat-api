package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// CORSMiddleware 创建 CORS 中间件（完整解决方案）
func CORSMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 动态设置允许的来源
			origin := r.Header.Get("Origin")

			// 允许的来源列表
			allowedOrigins := []string{
				"http://localhost:5173",
				"http://localhost:3000",
				"http://127.0.0.1:5173",
				"http://127.0.0.1:3000",
			}

			// 检查请求来源是否在允许列表中
			allowOrigin := ""
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					allowOrigin = origin
					break
				}
			}

			// 如果没有匹配，默认允许 localhost:5173（开发环境）
			if allowOrigin == "" {
				allowOrigin = "http://localhost:5173"
			}

			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24小时

			// 关键：正确处理 OPTIONS 预检请求
			if r.Method == "OPTIONS" {
				// 立即返回 204 无内容响应，不调用后续处理程序
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 对于非 OPTIONS 请求，继续后续处理
			next(w, r)
		}
	}
}
