package websocket

import (
	"math"
	"time"
)

// 默认属性
const (
	infinity = time.Duration(math.MaxInt64)

	defaultMaxConnectionIdle = infinity
)
