package middleware

import (
	"github.com/zeromicro/go-zero/rest"
	"net/http"
)

// CORSMiddleware 创建 CORS 中间件（完整解决方案）
func CORSMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 动态设置允许的来源
			origin := r.Header.Get("Origin")
			//allowedOrigin := "http://localhost:5173" // 默认只允许前端开发环境

			// 在生产环境中应该基于配置允许来源
			// 这里简化处理，实际项目中应该从配置中读取
			if origin == "http://localhost:5173" || origin == "https://your-production-domain.com" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
