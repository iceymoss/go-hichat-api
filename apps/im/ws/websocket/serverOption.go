package websocket

import (
	"github.com/iceymoss/go-hichat-api/pkg/presence"
	"time"
)

type Options func(opt *option)

type option struct {
	Authentication
	pattern string

	// ack 类型
	ack AckType

	// 因为在ack的机制中可能客户端一直未返回ack的信息，
	// 顾可以设置ackTimeout，如果超过这个时间说明目标接收消息有误可计入日志，
	// 以程序陷入死循环
	ackTimeout time.Duration

	// sendErrCount 发送失败计数
	sendErrCount int

	// 最大并发数
	maxConnectionIdle time.Duration

	concurrency     int
	nodeID          string
	presence        *presence.Store
	presenceTTL     time.Duration
	presenceRefresh time.Duration
}

func newOption(opts ...Options) option {
	o := option{
		pattern:           "/ws",
		ackTimeout:        defaultAckTimeout,
		sendErrCount:      sendErrCount,
		maxConnectionIdle: defaultMaxConnectionIdle,
		concurrency:       defaultConcurrency,
		presenceTTL:       5 * time.Minute,
		presenceRefresh:   2 * time.Minute,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func WithPresence(nodeID string, store *presence.Store, ttl, refresh time.Duration) Options {
	return func(opt *option) {
		opt.nodeID = nodeID
		opt.presence = store
		if ttl > 0 {
			opt.presenceTTL = ttl
		}
		if opt.presenceTTL < time.Second {
			opt.presenceTTL = time.Second
		}
		if refresh > 0 {
			opt.presenceRefresh = refresh
		}
		if opt.presenceRefresh >= opt.presenceTTL {
			opt.presenceRefresh = opt.presenceTTL / 2
		}
		if opt.presenceRefresh < time.Millisecond {
			opt.presenceRefresh = time.Millisecond
		}
	}
}

func WithMaxConnectionIdle(timeout time.Duration) Options {
	return func(opt *option) {
		if timeout > 0 {
			opt.maxConnectionIdle = timeout
		}
	}
}

// WithAck 设置ack相关属性
func WithAck(ack AckType) Options {
	return func(opt *option) {
		opt.ack = ack
	}
}

func WithConcurrency(concurrency int) Options {
	return func(opt *option) {
		opt.concurrency = concurrency
	}
}
func WithAuthentication(authentication Authentication) Options {
	return func(opt *option) {
		opt.Authentication = authentication
		opt.ack = RigorAck
	}
}

func WithHandlerPatten(pattern string) Options {
	return func(opt *option) {
		opt.pattern = pattern
	}
}
