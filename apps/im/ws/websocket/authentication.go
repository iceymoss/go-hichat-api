package websocket

import (
	"fmt"
	"net/http"
	"time"
)

type Authentication interface {
	// Auth 用户权限校验
	Auth(s *Server, w http.ResponseWriter, r *http.Request) bool

	// UserId 获取用户id
	UserId(r *http.Request) string
}

type authentication struct{}

func (*authentication) Auth(s *Server, w http.ResponseWriter, r *http.Request) bool {
	fmt.Println("token:", r.Header.Get("Authorization"))
	return true
}

func (*authentication) UserId(r *http.Request) string {
	query := r.URL.Query()
	if query != nil && query["userId"] != nil {
		return fmt.Sprintf("%v", query["userId"])
	}

	return fmt.Sprintf("%v", time.Now().UnixMilli())
}
