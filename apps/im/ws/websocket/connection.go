package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// Conn 对websocket连接进行一层封装，方便做功能拓展
type Conn struct {
	idleMu sync.Mutex

	Uid string

	// websocket 连接
	Conn *websocket.Conn

	// websocket 服务对象
	s *Server

	// 保证并发安全
	mu sync.Mutex

	// 连接的空闲时间
	idle time.Time

	// 运行最大空闲定时器
	maxConnectionIdle time.Duration

	messageMu      sync.Mutex
	readMessage    []*Message
	readMessageSeq map[string]*Message

	message chan *Message

	done chan struct{}
}

// NewConn 将连接转为websocket 并且对该连接进行进一步封装
func NewConn(s *Server, w http.ResponseWriter, r *http.Request) *Conn {
	c, err := s.upGrader.Upgrade(w, r, nil)
	if err != nil {
		s.Errorf("upgrade err %v", err)
		return nil
	}

	conn := &Conn{
		Conn:              c,
		s:                 s,
		idle:              time.Now(),
		readMessage:       make([]*Message, 0, 2),
		readMessageSeq:    make(map[string]*Message, 2),
		message:           make(chan *Message, 1),
		done:              make(chan struct{}),
		maxConnectionIdle: defaultMaxConnectionIdle,
	}

	return conn
}

// keepalive Conn 结构体的保活检测方法，用于监控连接空闲时间
func (c *Conn) keepalive() {
	// 初始化空闲超时定时器（maxConnectionIdle 为允许的最大空闲时长）
	idleTimer := time.NewTimer(c.maxConnectionIdle)

	// 确保退出时释放定时器资源
	defer idleTimer.Stop()

	for {
		select {
		// 情况1：连接空闲超时定时器触发
		case <-idleTimer.C:
			c.mu.Lock()

			// 获取连接的最后活跃时间
			idle := c.idle

			fmt.Printf("idle %v, maxIdle %v \n", c.idle, c.maxConnectionIdle)
			if idle.IsZero() { // The connection is non-idle.
				c.mu.Unlock()
				// 重置定时器，继续监控下一个周期
				idleTimer.Reset(c.maxConnectionIdle)
				continue
			}

			// 计算剩余允许空闲时间 = 最大空闲时间 - 已空闲时间
			val := c.maxConnectionIdle - time.Since(idle)
			fmt.Printf("val %v \n", val)

			// 如果最后活跃时间为零值（表示连接处于活跃状态）
			c.mu.Unlock()

			// 如果连接空闲时间超过阈值
			if val <= 0 {
				// 优雅关闭连接（如发送 FIN 包）
				// The connection has been idle for a duration of keepalive.MaxConnectionIdle or more.
				// Gracefully close the connection.
				c.s.Close(c)
				return
			}
			// 重置定时器，等待剩余允许空闲时间
			idleTimer.Reset(val)

		// 情况2：收到连接终止信号（通过关闭 done 通道）
		case <-c.done:
			fmt.Println("客户端结束连接")
			return
		}
	}
}

// ReadMessage 从客户端读取到消息，表示不空闲
func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	// 开始忙碌
	messageType, p, err = c.Conn.ReadMessage()
	c.idle = time.Time{}
	return
}

// WriteMessage 当我写入数据后，就被定义为空闲时间了
func (c *Conn) WriteMessage(messageType int, data []byte) error {
	err := c.Conn.WriteMessage(messageType, data)
	// 当写操作完成后当前连接就会进入空闲状态，并记录空闲的时间
	c.idle = time.Now()
	return err
}

func (c *Conn) Close() error {
	close(c.done)
	return c.Conn.Close()
}
