package websocket

import (
	"math"
	"time"
)

// 默认属性
const (
	infinity = time.Duration(math.MaxInt64)

	defaultMaxConnectionIdle = infinity

	// defaultAckTimeout ack的默认最长确认时间
	defaultAckTimeout = 30 * time.Second

	sendErrCount = 5
)
