package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CodeMsg struct {
	Code int    // 错误码
	Msg  string // 错误消息
	Err  error  // 原始错误
}

// 实现 error 接口
func (e *CodeMsg) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", e.Code, e.Msg)
}

// GRPCStatus 实现 gRPC 状态转换接口
func (e *CodeMsg) GRPCStatus() *status.Status {
	return status.New(codes.Code(e.Code), e.Msg)
}

// New 构造函数
func New(code int, msg string) error {
	return &CodeMsg{Code: code, Msg: msg}
}
