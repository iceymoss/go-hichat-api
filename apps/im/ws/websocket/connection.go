package websocket

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Conn 对websocket连接进行一层封装，方便做功能拓展
type Conn struct {
	idleMu sync.Mutex

	// 当前连接的用户id
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

	// 消息确认机制的锁
	messageMu sync.Mutex

	// 当前连接的消息确认队列, 接收消息处理队列，采用数组存储可保障数据的顺序。
	readMessages []*Message

	// 当前连接的消息, 记录ack机制中的消息处理结果与进展
	readMessageSeq map[string]*Message

	// 用于和writeHandler协程进行通信,消息通道，在ack验证完成后将消息投递与writeHandler处理。
	message chan *Message

	// 当前连接状态
	done      chan struct{}
	closeOnce sync.Once
	writeMu   sync.Mutex
}

// NewConn 将连接转为websocket 并且对该连接进行进一步封装
func NewConn(s *Server, w http.ResponseWriter, r *http.Request) *Conn {
	c, err := s.upGrader.Upgrade(w, r, nil)
	if err != nil {
		s.Errorf("upgrade err %v", err)
		return nil
	}

	uid := s.authentication.UserId(r)
	fmt.Println("连接的uid", uid)
	//需要满足三次ack操作，客户端发起-> 服务端收到消息并且确认消息 -> 客户端收到消息确认消息 -> ack成功，服务端做后续处理
	conn := &Conn{
		Conn:              c,
		s:                 s,
		Uid:               uid,
		idle:              time.Now(),
		readMessages:      make([]*Message, 0, 2),       // 记录客户端的两次消息
		readMessageSeq:    make(map[string]*Message, 2), // 记录客户端的两次消息
		message:           make(chan *Message, 1),       // 用于同步消息给WriteHandler的chan,容量为1刚刚好使得channel是有缓存，但因为容量只有1可以保证发送和接收操作的顺序，从而确保数据的有序性。
		done:              make(chan struct{}),
		maxConnectionIdle: s.opt.maxConnectionIdle,
	}
	c.SetPongHandler(func(string) error { conn.idleMu.Lock(); conn.idle = time.Now(); conn.idleMu.Unlock(); return nil })

	go conn.keepalive()

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
			c.idleMu.Lock()

			// 获取连接的最后活跃时间
			idle := c.idle

			fmt.Printf("idle %v, maxIdle %v \n", c.idle, c.maxConnectionIdle)
			if idle.IsZero() { // The connection is non-idle.
				c.idleMu.Unlock()
				// 重置定时器，继续监控下一个周期
				idleTimer.Reset(c.maxConnectionIdle)
				continue
			}

			// 计算剩余允许空闲时间 = 最大空闲时间 - 已空闲时间
			val := c.maxConnectionIdle - time.Since(idle)
			fmt.Printf("val %v \n", val)

			// 如果最后活跃时间为零值（表示连接处于活跃状态）
			c.idleMu.Unlock()

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

// 将消息记录到队列中，一个消息的确认流程，消息id是唯一的
func (c *Conn) appendMsgMq(msg *Message) {
	c.messageMu.Lock()
	defer c.messageMu.Unlock()
	// 读队列
	if m, ok := c.readMessageSeq[msg.Id]; ok {
		// 客户端可能重复发送了消息 或者 收到ack消息
		if len(c.readMessages) == 0 {
			// 数据已经被处理不存在，顾属于重复了
			return
		}

		//m.Id != msg.Id 不是同一个消息的ack, 或者m.AckSeq >= msg.AckSeq 都说明，消息重复了
		if m.Id != msg.Id || m.AckSeq >= msg.AckSeq {
			// 数据已经被处理不存在，顾属于重复了
			// ack的消息内容还是一致或大于接收的序号说明也是重复了
			return
		}

		// 等于最新的记录，记录消息确认最新状态
		c.readMessageSeq[msg.Id] = msg
		fmt.Printf("消息被客户端确认了: %+v\n", msg)
		return
	}

	// 注意：后续逻辑只能是客户端某一个消息的第一次发送
	// 客户端不能一上来就发送FrameAck类型的消息
	// 这个只能用来做应答的，不应该出现在之类
	// 因为意外发送ack信息，直接过滤
	if msg.FrameType == FrameAck {
		return
	}

	// 记录该消息
	c.readMessages = append(c.readMessages, msg)
	c.readMessageSeq[msg.Id] = msg
}

// ReadMessage 从客户端读取到消息，表示不空闲
func (c *Conn) ReadMessage() (messageType int, p []byte, err error) {
	// 开始忙碌
	messageType, p, err = c.Conn.ReadMessage()
	if err == nil {
		c.idleMu.Lock()
		c.idle = time.Now()
		c.idleMu.Unlock()
	}
	return
}

// WriteMessage 当我写入数据后，就被定义为空闲时间了
func (c *Conn) WriteMessage(messageType int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.s.opt.writeTimeout > 0 {
		_ = c.Conn.SetWriteDeadline(time.Now().Add(c.s.opt.writeTimeout))
	}
	err := c.Conn.WriteMessage(messageType, data)
	return err
}

func (c *Conn) Close() error {
	var err error
	c.closeOnce.Do(func() { close(c.done); err = c.Conn.Close() })
	return err
}
