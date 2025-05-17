package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

// Server websocket服务对象
type Server struct {
	// 并发安全处理
	sync.RWMutex

	// 监听地址
	addr string

	// websocket相关
	upGrader websocket.Upgrader

	// 绑定的处理方法
	routes map[string]HandlerFunc

	// 用户绑定连接
	user2Conn map[string]*websocket.Conn

	// 连接绑定用户
	conn2User map[*websocket.Conn]string

	// 用户权限相关
	authentication Authentication

	// 可选项
	opt Options

	// 日志相关
	logx.Logger
}

// NewServer 初始化一个websocket服务实例
func NewServer(addr string, opt ...Options) *Server {
	return &Server{
		addr:           addr,
		upGrader:       websocket.Upgrader{},
		routes:         make(map[string]HandlerFunc),
		user2Conn:      make(map[string]*websocket.Conn),
		conn2User:      make(map[*websocket.Conn]string),
		authentication: newOption(opt...),
		Logger:         logx.WithContext(context.Background()),
	}
}

// ServerWs websocket服务对外暴露入口
func (s *Server) ServerWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.Errorf("server handler ws recover err %v", r)
		}
	}()

	// 权限校验
	if !s.authentication.Auth(w, r) {
		s.Info("authentication failed")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 向请求变为websocket
	conn, err := s.upGrader.Upgrade(w, r, nil)
	if err != nil {
		s.Error("upgrade http conn err", err)
		return
	}

	// 添加用户和连接间的映射
	s.addConn(conn, r)

	go s.handlerConn(conn)
}

// handlerConn 监听连接上的消息
func (s *Server) handlerConn(conn *websocket.Conn) {
	// 记录连接
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.Errorf("websocket conn readMessage err %v, user Id %s", err, "")
			// 关闭并删除连接
			s.Close(conn)
			return
		}

		// 请求信息
		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			zLog.Error("handlerConn.Unmarshal: unmarshal failed ", zap.Error(err))
			s.Close(conn)
			return
		}

		// 获取处理方法
		if handler, ok := s.routes[message.Method]; ok {
			handler(s, conn, &message)
		} else {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("不存在请求方法 %v 请仔细检查", message.Method)))
		}
	}
}

// SendByUserId 向指定用户发送消息
func (s *Server) SendByUserId(msg interface{}, sendIds ...string) error {
	if len(sendIds) == 0 {
		return nil
	}

	conn := s.GetConn(sendIds)

	return s.Send(msg, conn...)
}

// Send push消息到一个或者多个conn中
func (s *Server) Send(msg interface{}, conns ...*websocket.Conn) error {
	if len(conns) == 0 {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for _, conn := range conns {
		if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return err
		}
	}
	return nil
}

// AddRoutes 为方法添加handler
func (s *Server) AddRoutes(rs []Route) {
	for _, r := range rs {
		s.routes[r.Method] = r.Handler
	}
}

// addConn 添加连接和用户的映射
func (s *Server) addConn(conn *websocket.Conn, req *http.Request) {
	uid := s.authentication.UserId(req)

	fmt.Println("连接池:", len(s.user2Conn))
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	s.user2Conn[uid] = conn
	s.conn2User[conn] = uid

	fmt.Println("连接池变化:", len(s.user2Conn))
}

func (s *Server) GetConn(uids []string) []*websocket.Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	if len(uids) == 0 {
		res := make([]*websocket.Conn, 0, len(s.user2Conn))
		for _, conn := range s.user2Conn {
			res = append(res, conn)
		}
		return res
	}

	res := make([]*websocket.Conn, 0, len(uids))
	for _, uid := range uids {
		if conn, ok := s.user2Conn[uid]; ok {
			res = append(res, conn)
		}
	}

	return res
}

// GetUsers 根据连接获取用户id
func (s *Server) GetUsers(conns []*websocket.Conn) []string {
	s.RWMutex.RLock()
	defer s.RUnlock()

	if len(conns) == 0 {
		// get all
		res := make([]string, 0, len(s.conn2User))
		for _, uid := range s.conn2User {
			res = append(res, uid)
		}
		return res
	}

	res := make([]string, 0, len(conns))
	for _, conn := range conns {
		if uid, ok := s.conn2User[conn]; ok {
			res = append(res, uid)
		}
	}
	return res
}

// Start 开始websocket服务
func (s *Server) Start() {
	http.HandleFunc("/ws", s.ServerWs)
	http.ListenAndServe(s.addr, nil)
}

// Stop 终止websocket服务
func (s *Server) Stop() {
	fmt.Println("stop server")
}

// Close 关闭连接
func (s *Server) Close(conn *websocket.Conn) {
	conn.Close()

	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	fmt.Println("退出连接池:", len(s.user2Conn))

	// 关闭连接需要维护"连接-用户"关系
	uid := s.conn2User[conn]
	delete(s.conn2User, conn)
	delete(s.user2Conn, uid)

	fmt.Println("退出连接池变化", len(s.conn2User))
}
