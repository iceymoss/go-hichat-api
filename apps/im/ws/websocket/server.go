package websocket

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"go.uber.org/zap"
)

type AckType int

const (
	// NoAck 不进行ack确认
	NoAck AckType = iota
	// OnlyAck 只回 - 两次通信
	OnlyAck
	// RigorAck 严格 - 三次通信
	RigorAck
)

func (t AckType) ToString() string {
	switch t {
	case OnlyAck:
		return "OnlyAck"
	case RigorAck:
		return "RigorAck"
	}
	return "NoAck"
}

// Server websocket服务对象
type Server struct {
	// 并发安全处理
	sync.RWMutex
	presenceMu [256]sync.Mutex

	// 监听地址
	addr string

	// websocket相关
	upGrader websocket.Upgrader

	// 绑定的处理方法
	routes map[string]HandlerFunc

	// 用户绑定连接
	user2Conn map[string]map[*Conn]struct{}

	// 连接绑定用户
	conn2User      map[*Conn]string
	presenceLeases map[string]*userPresenceLease

	// 用户权限相关
	authentication Authentication

	// 可选项
	opt option

	// 日志相关
	logx.Logger

	// 处理并发数
	TaskRunner *threading.TaskRunner
}

type userPresenceLease struct {
	token string
	done  chan struct{}
	once  sync.Once
}

func (l *userPresenceLease) stop() { l.once.Do(func() { close(l.done) }) }

// NewServer 初始化一个websocket服务实例
func NewServer(addr string, opt ...Options) *Server {
	return &Server{
		addr: addr,
		upGrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		routes:         make(map[string]HandlerFunc),
		user2Conn:      make(map[string]map[*Conn]struct{}),
		conn2User:      make(map[*Conn]string),
		presenceLeases: make(map[string]*userPresenceLease),
		authentication: newOption(opt...),
		opt:            newOption(opt...),
		Logger:         logx.WithContext(context.Background()),
		TaskRunner:     threading.NewTaskRunner(newOption(opt...).concurrency),
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

	// 构造websocket连接
	conn := NewConn(s, w, r)
	if conn == nil {
		zLog.Error("ServerWs.NewConn: build websocket failed")
		return
	}

	// 添加用户和连接间的映射
	s.addConn(conn, r)

	go s.handlerConn(conn)
}

// handlerConn 监听连接上的消息
func (s *Server) handlerConn(conn *Conn) {

	// 协程处理ack完成后的逻辑
	go s.handleWrite(conn)

	fmt.Println("系统ack级别：", s.opt.ack)

	// 系统ack级别，如果需要ack
	if s.opt.ack != NoAck {
		// 接收ack并且确认
		go s.readAck(conn)
	}

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
			zLog.Error("handlerConn.Unmarshal: unmarshal failed ", zap.Any("msg", string(msg)), zap.Error(err))
			s.Send(NewErrMessage(err), conn)
			//s.Close(conn)
			continue
		}

		// TODO: 这里有一个bug，当用户直接第一次传入一个ack = 2的值，会导致当前用户连接整个ack机制锁死，消息发送不了了

		fmt.Printf("socket服务器收到消息: %+v", message)
		if s.opt.ack != NoAck && message.FrameType != FrameNoAck {
			// 将消息写入队列
			conn.appendMsgMq(&message)
		} else {
			// 直接将消息交给handleWrite处理
			conn.message <- &message
		}
	}
}

func (s *Server) readAck(conn *Conn) {

	send := func(msg *Message, conn *Conn) error {
		fmt.Printf("发送消息给客户端: %+v\n", msg)
		err := s.Send(msg, conn)
		if err == nil {
			return nil
		}

		s.Errorf("message ack OnlyAck send err %v message %v", err, msg)
		conn.readMessages[0].errCount++
		conn.messageMu.Unlock()

		tempDelay := time.Duration(200*conn.readMessages[0].errCount) * time.Microsecond
		if max := 1 * time.Second; tempDelay > max {
			tempDelay = max
		}

		time.Sleep(tempDelay)
		return err
	}

	for {
		select {
		case <-conn.done:
			// 关闭了连接
			s.Infof("close message ack uid %v", conn.Uid)
			return
		default:
		}

		conn.messageMu.Lock()
		// 这里相当于自旋
		if len(conn.readMessages) == 0 {
			conn.messageMu.Unlock()
			// 队列为空时短暂等待，避免 CPU 空转
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 取出队列中的第一个数据
		message := conn.readMessages[0]
		if message.errCount > s.opt.sendErrCount {
			s.Infof("conn send fail, message %v, ackType %v, maxSendErrCount %v", message, s.opt.ack.ToString(), s.opt.sendErrCount)
			conn.messageMu.Unlock()
			// todo：因为发送消息多次错误，而选择放弃消息
			delete(conn.readMessageSeq, message.Id)
			conn.readMessages = conn.readMessages[1:]
			continue
		}

		fmt.Printf("ack应答中 message: %+v\n", message)

		// 根据ack的确认策略选择合适的处理方式
		switch s.opt.ack {
		case OnlyAck:
			if err := send(&Message{
				FrameType: FrameAck,
				AckSeq:    message.AckSeq + 1,
				Id:        message.Id,
			}, conn); err != nil {
				continue
			}
			// 只回答, 向客户端发送ack
			conn.readMessages = conn.readMessages[1:]
			conn.messageMu.Unlock()
			conn.message <- message
			s.Infof("message ack OnlyAck send success mid %v", message.Id)
		case RigorAck:
			if message.AckSeq == 0 {
				// 还未发送过确认信息
				conn.readMessages[0].AckSeq++
				conn.readMessages[0].ackTime = time.Now()
				if err := send(&Message{
					FrameType: FrameAck,
					AckSeq:    message.AckSeq,
					Id:        message.Id,
				}, conn); err != nil {
					continue
				}

				conn.messageMu.Unlock()
				s.Infof("message ack RigorAck send mid %v, seq %v, time %v", message.Id, message.AckSeq, message.ackTime.Unix())
				continue
			}

			// 已经发送过序号了，需要等待客户端返回确认
			msgSeq := conn.readMessageSeq[message.Id]
			if msgSeq.AckSeq > message.AckSeq {
				// 客户端确认成功,可以处理业务了
				conn.readMessages = conn.readMessages[1:]
				conn.messageMu.Unlock()
				conn.message <- message
				s.Infof("message ack RigorAck success mid %v ack-type %v ", message.Id, message.FrameType)
				continue
			}

			// 很显然没有处理成功，先看看有没有超时
			val := s.opt.ackTimeout - time.Since(message.ackTime)

			if !message.ackTime.IsZero() && val <= 0 {
				// todo: 超时了，可以选择断开与客户端的连接,但实际具体细节处理仍然还需自己结合业务完善，此处选择放弃该消息
				s.Errorf("message ack RigorAck fail mid %v, time %v because timeout", message.Id, message.ackTime)
				delete(conn.readMessageSeq, message.Id)
				conn.readMessages = conn.readMessages[1:]
				conn.messageMu.Unlock()
				continue
			}

			conn.messageMu.Unlock()
			if val > 0 && val > 300*time.Microsecond {
				if err := send(&Message{
					FrameType: FrameAck,
					AckSeq:    message.AckSeq,
					Id:        message.Id,
				}, conn); err != nil {
					continue
				}
			}
			// 等待客户端 ACK 确认，短轮询
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// handleWrite ack完成后处理后续业务逻辑
func (s *Server) handleWrite(conn *Conn) {
	for {
		select {
		case <-conn.done:
			// 连接被关闭，结束
			return
		case message := <-conn.message: // 监听conn.message的消息
			// 依据请求消息类型分类处理
			fmt.Printf("handleWrite 收到消息，准备推入mq：%+v", message)
			switch message.FrameType {
			case FramePing:
				// ping：回复
				s.Send(&Message{FrameType: FramePing}, conn)
			case FrameData, FrameNoAck: // 普通消息或者不做消息ack
				// 这里调用真正的逻辑处理后续业务
				if handler, ok := s.routes[message.Method]; ok {
					handler(s, conn, message)
				} else {
					s.Send(&Message{
						FrameType: FrameData,
						Data:      fmt.Sprintf("不存在请求方法 %v 请仔细检查", message.Method),
					}, conn)
				}
			}

			if s.opt.ack != NoAck {
				// 删除 消息ack的序号记录
				conn.messageMu.Lock()
				delete(conn.readMessageSeq, message.Id)
				conn.messageMu.Unlock()
			}
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
func (s *Server) Send(msg interface{}, conns ...*Conn) error {
	if len(conns) == 0 {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var errs []error
	failed := make([]*Conn, 0)
	for _, conn := range conns {
		if conn == nil {
			continue
		}
		if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
			errs = append(errs, err)
			failed = append(failed, conn)
		}
	}
	for _, conn := range failed {
		go s.Close(conn)
	}
	return errors.Join(errs...)
}

// AddRoutes 为方法添加handler
func (s *Server) AddRoutes(rs []Route) {
	for _, r := range rs {
		s.routes[r.Method] = r.Handler
	}
}

// addConn 添加连接和用户的映射
func (s *Server) addConn(conn *Conn, req *http.Request) {
	uid := s.authentication.UserId(req)
	if uid == "" {
		conn.Close()
		return
	}
	presenceLock := s.presenceLock(uid)
	presenceLock.Lock()
	defer presenceLock.Unlock()
	s.RWMutex.Lock()

	connections := s.user2Conn[uid]
	first := len(connections) == 0
	if first {
		connections = make(map[*Conn]struct{})
		s.user2Conn[uid] = connections
	}
	connections[conn] = struct{}{}
	s.conn2User[conn] = uid
	if !first || s.opt.presence == nil || uid == "101" {
		s.RWMutex.Unlock()
		return
	}
	if s.opt.presence != nil {
		tokenBytes := make([]byte, 16)
		if _, err := rand.Read(tokenBytes); err != nil {
			delete(connections, conn)
			delete(s.conn2User, conn)
			if len(connections) == 0 {
				delete(s.user2Conn, uid)
			}
			s.RWMutex.Unlock()
			conn.Close()
			s.Errorf("generate presence token uid=%s: %v", uid, err)
			return
		}
		lease := &userPresenceLease{token: hex.EncodeToString(tokenBytes), done: make(chan struct{})}
		s.presenceLeases[uid] = lease
		s.RWMutex.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := s.opt.presence.Claim(ctx, uid, s.opt.nodeID, lease.token, s.opt.presenceTTL)
		cancel()
		claimed := err == nil
		if err != nil {
			s.Errorf("claim websocket presence uid=%s: %v", uid, err)
		}
		go s.refreshPresence(lease, uid, claimed)
	}
}

func (s *Server) refreshPresence(lease *userPresenceLease, uid string, claimed bool) {
	ticker := time.NewTicker(s.opt.presenceRefresh)
	defer ticker.Stop()
	for {
		select {
		case <-lease.done:
			return
		case <-ticker.C:
			presenceLock := s.presenceLock(uid)
			presenceLock.Lock()
			s.RWMutex.RLock()
			current := s.presenceLeases[uid] == lease && len(s.user2Conn[uid]) > 0
			s.RWMutex.RUnlock()
			if !current {
				presenceLock.Unlock()
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if !claimed {
				err := s.opt.presence.Claim(ctx, uid, s.opt.nodeID, lease.token, s.opt.presenceTTL)
				cancel()
				if err != nil {
					s.Errorf("claim websocket presence uid=%s: %v", uid, err)
					presenceLock.Unlock()
					continue
				}
				claimed = true
				presenceLock.Unlock()
				continue
			}
			owned, err := s.opt.presence.Refresh(ctx, uid, lease.token, s.opt.presenceTTL)
			cancel()
			if err != nil {
				s.Errorf("refresh websocket presence uid=%s: %v", uid, err)
				presenceLock.Unlock()
				continue
			}
			if !owned {
				presenceLock.Unlock()
				return
			}
			presenceLock.Unlock()
		}
	}
}

func (s *Server) GetConn(uids []string) []*Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	if len(uids) == 0 {
		res := make([]*Conn, 0, len(s.conn2User))
		for _, connections := range s.user2Conn {
			for conn := range connections {
				res = append(res, conn)
			}
		}
		return res
	}

	res := make([]*Conn, 0, len(uids))
	seen := make(map[*Conn]struct{})
	for _, uid := range uids {
		for conn := range s.user2Conn[uid] {
			if _, ok := seen[conn]; !ok {
				seen[conn] = struct{}{}
				res = append(res, conn)
			}
		}
	}

	return res
}

// GetUsers 根据连接获取用户id
func (s *Server) GetUsers(conns []*Conn) []string {
	s.RWMutex.RLock()
	defer s.RUnlock()

	if len(conns) == 0 {
		// get all
		res := make([]string, 0, len(s.user2Conn))
		for uid := range s.user2Conn {
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
func (s *Server) Close(conn *Conn) {
	presenceLock := s.presenceLock(conn.Uid)
	presenceLock.Lock()
	defer presenceLock.Unlock()
	s.RWMutex.Lock()

	// 避免出现重复关闭问题
	uid := s.conn2User[conn]
	if uid == "" {
		s.RWMutex.Unlock()
		conn.Close()
		return
	}

	delete(s.conn2User, conn)
	connections := s.user2Conn[uid]
	delete(connections, conn)
	last := len(connections) == 0
	var lease *userPresenceLease
	if last {
		delete(s.user2Conn, uid)
		lease = s.presenceLeases[uid]
		delete(s.presenceLeases, uid)
		if lease != nil {
			lease.stop()
		}
	}
	s.RWMutex.Unlock()

	conn.Close()
	if last && lease != nil && s.opt.presence != nil && uid != "101" {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := s.opt.presence.DeleteIfOwner(ctx, uid, lease.token)
		cancel()
		if err != nil {
			s.Errorf("delete websocket presence uid=%s: %v", uid, err)
		}
	}

}

func (s *Server) presenceLock(uid string) *sync.Mutex {
	h := fnv.New32a()
	_, _ = h.Write([]byte(uid))
	return &s.presenceMu[h.Sum32()%uint32(len(s.presenceMu))]
}
