package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client websocket 的客户端
type Client interface {
	Close() error
	Send(v any) error
	Read(v any) error
}

type client struct {
	*websocket.Conn
	host        string
	opt         dailOption
	mu          sync.Mutex
	stateMu     sync.Mutex
	closed      bool
	dialCtx     context.Context
	cancelDial  context.CancelFunc
	dialContext func(context.Context, string, http.Header) (*websocket.Conn, *http.Response, error)
}

var errClientClosed = errors.New("websocket client is closed")

func NewClient(host string, opts ...DailOptions) *client {
	opt := newDailOptions(opts...)
	dialCtx, cancelDial := context.WithCancel(context.Background())

	c := &client{
		opt:         opt,
		host:        host,
		dialCtx:     dialCtx,
		cancelDial:  cancelDial,
		dialContext: websocket.DefaultDialer.DialContext,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := c.dail()
	if err != nil {
		panic(err)
	}
	c.Conn = conn
	return c
}

// dail 连接服务端
func (c *client) dail() (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: c.host, Path: c.opt.pattern}
	fmt.Println("url:", u.String())
	ctx := c.dialCtx
	if ctx == nil {
		ctx = context.Background()
	}
	dial := c.dialContext
	if dial == nil {
		dial = websocket.DefaultDialer.DialContext
	}
	conn, _, err := dial(ctx, u.String(), c.opt.header)
	return conn, err
}

// Send 给websocket发送数据
func (c *client) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		zLog.Error("Send.Marshal: json marshal failed", zap.Any("msg", string(data)), zap.Error(err))
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed() {
		return errClientClosed
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err == nil {
		zLog.Info("Send.WriteMessage: push to websocket succeed", zap.Any("message", data))
		return nil
	}

	zLog.Error("Send.WriteMessage: push data to websocket serve failed", zap.Any("msg", string(data)), zap.Error(err))

	// 发送失败了再建立一次连接
	if c.isClosed() {
		return errClientClosed
	}
	conn, cerr := c.dail()
	if cerr != nil {
		zLog.Error("Send.dail: dail websocket serve failed", zap.Any("message", data), zap.Error(err))
		return err
	}
	c.stateMu.Lock()
	if c.closed {
		c.stateMu.Unlock()
		_ = conn.Close()
		return errClientClosed
	}
	c.Conn = conn
	c.stateMu.Unlock()
	return c.WriteMessage(websocket.TextMessage, data)
}

func (c *client) Close() error {
	c.stateMu.Lock()
	if c.closed {
		c.stateMu.Unlock()
		return nil
	}
	c.closed = true
	conn := c.Conn
	if c.cancelDial != nil {
		c.cancelDial()
	}
	c.stateMu.Unlock()
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func (c *client) isClosed() bool {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	return c.closed
}

// Read 读取websocket数据
func (c *client) Read(v any) error {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}
