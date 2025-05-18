package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
)

// Client websocket 的客户端
type Client interface {
	Close() error
	Send(v any) error
	Read(v any) error
}

type client struct {
	*websocket.Conn
	host string
	opt  dailOption
}

func NewClient(host string, opts ...DailOptions) *client {
	opt := newDailOptions(opts...)

	c := &client{
		opt:  opt,
		host: host,
	}
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
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), c.opt.header)
	return conn, err
}

// Send 给websocket发送数据
func (c *client) Send(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.TextMessage, data)
	if err == nil {
		return nil
	}

	// 发送失败了再建立一次连接
	conn, cerr := c.dail()
	if cerr != nil {
		return err
	}
	c.Conn = conn
	return c.WriteMessage(websocket.TextMessage, data)
}

// Read 读取websocket数据
func (c *client) Read(v any) error {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}
