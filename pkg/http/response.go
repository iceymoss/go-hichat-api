package http

import (
	"context"
	"net/http"

	"github.com/iceymoss/go-hichat-api/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	zrpcErr "github.com/zeromicro/x/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	ErrorCode string      `json:"error_code,omitempty"`
}

func Success(data interface{}) *Response {
	return &Response{
		Code: 200,
		Msg:  "",
		Data: data,
	}
}

func Fail(code int, err string) *Response {
	return &Response{
		Code: code,
		Msg:  err,
		Data: nil,
	}
}

func OkHandler(_ context.Context, v interface{}) any {
	return Success(v)
}

func ErrHandler(name string) func(ctx context.Context, err error) (int, any) {
	return func(ctx context.Context, err error) (int, any) {
		errcode := xerr.SERVER_COMMON_ERROR
		errmsg := xerr.ErrMsg(errcode)

		causeErr := errors.Cause(err)
		if e, ok := causeErr.(*zrpcErr.CodeMsg); ok {
			errcode = e.Code
			errmsg = e.Msg
		} else {
			if gstatus, ok := status.FromError(causeErr); ok {
				errcode = int(gstatus.Code())
				errmsg = gstatus.Message()
			}
		}

		// 日志记录
		logx.WithContext(ctx).Errorf("【%s】 err %v", name, err)

		response := Fail(errcode, errmsg)
		if _, ok := causeErr.(*zrpcErr.CodeMsg); !ok {
			response.ErrorCode = stableErrorCode(causeErr)
		}
		return errorHTTPStatus(causeErr), response
	}
}

func stableErrorCode(err error) string {
	if s, ok := status.FromError(err); ok {
		for _, detail := range s.Details() {
			if info, ok := detail.(*errdetails.ErrorInfo); ok && info.Reason != "" {
				return info.Reason
			}
		}
	}
	switch status.Code(err) {
	case codes.InvalidArgument:
		return "invalid_argument"
	case codes.Unauthenticated:
		return "unauthenticated"
	case codes.PermissionDenied:
		return "forbidden"
	case codes.NotFound:
		return "not_found"
	case codes.AlreadyExists, codes.Aborted, codes.FailedPrecondition:
		return "conflict"
	case codes.ResourceExhausted:
		return "rate_limited"
	case codes.Internal, codes.DataLoss:
		return "internal"
	default:
		return "unknown"
	}
}

func errorHTTPStatus(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted, codes.FailedPrecondition:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Internal, codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
