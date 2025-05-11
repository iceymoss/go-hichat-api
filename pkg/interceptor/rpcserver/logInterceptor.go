package rpcserver

import (
	"context"
	zLog "github.com/iceymoss/go-hichat-api/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	resp, err = handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	//logx.WithContext(ctx).Errorf("【RPC SRV ERR】 %+v", err)
	zLog.Error("【RPC SRV ERR】:", zap.Any("err:", err))

	// 直接转换为 gRPC 状态错误（利用 GRPCStatus 接口）
	if st, ok := status.FromError(err); ok {
		return resp, st.Err()
	}

	// 通用错误处理
	return resp, status.Error(codes.Unknown, err.Error())
}
