package websocket

import "github.com/gorilla/websocket"

type Route struct {
	Method  string
	Handler HandlerFunc
}

type HandlerFunc func(srv *Server, conn *websocket.Conn, msg *Message)
