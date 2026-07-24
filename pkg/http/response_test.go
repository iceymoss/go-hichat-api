package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeromicro/x/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrHandlerPreservesEnvelope(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		httpStatus int
		code       int
		message    string
		errorCode  string
	}{
		{name: "invalid argument", err: status.Error(codes.InvalidArgument, "bad input"), httpStatus: http.StatusBadRequest, code: int(codes.InvalidArgument), message: "bad input", errorCode: "invalid_argument"},
		{name: "forbidden", err: status.Error(codes.PermissionDenied, "forbidden"), httpStatus: http.StatusForbidden, code: int(codes.PermissionDenied), message: "forbidden", errorCode: "forbidden"},
		{name: "conflict", err: status.Error(codes.FailedPrecondition, "handled"), httpStatus: http.StatusConflict, code: int(codes.FailedPrecondition), message: "handled", errorCode: "conflict"},
		{name: "business code", err: errors.New(1200, "business forbidden"), httpStatus: http.StatusBadRequest, code: 1200, message: "business forbidden"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpStatus, body := ErrHandler("test")(context.Background(), test.err)
			require.Equal(t, test.httpStatus, httpStatus)
			resp, ok := body.(*Response)
			require.True(t, ok)
			require.Equal(t, test.code, resp.Code)
			require.Equal(t, test.message, resp.Msg)
			require.Nil(t, resp.Data)
			if test.errorCode != "" {
				require.Equal(t, test.errorCode, resp.ErrorCode)
			}
		})
	}
}
