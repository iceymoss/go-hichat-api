package logic

import (
	"context"
	"errors"

	"github.com/iceymoss/go-hichat-api/pkg/rpcauth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requireNotificationUser(ctx context.Context, auth *rpcauth.Auth, receiverID string) error {
	if auth == nil {
		return status.Error(codes.Unauthenticated, "missing rpc caller principal")
	}
	if err := rpcauth.RequireUser(ctx, receiverID); err != nil {
		if errors.Is(err, rpcauth.ErrWrongPrincipal) {
			return status.Error(codes.PermissionDenied, "rpc caller cannot access this receiver")
		}
		return status.Error(codes.Unauthenticated, "missing or invalid rpc caller principal")
	}
	return nil
}

func requireNotificationTask(ctx context.Context, auth *rpcauth.Auth) error {
	if auth == nil {
		return status.Error(codes.Unauthenticated, "missing rpc caller principal")
	}
	if err := rpcauth.RequireTask(ctx); err != nil {
		if errors.Is(err, rpcauth.ErrWrongPrincipal) {
			return status.Error(codes.PermissionDenied, "task rpc caller principal required")
		}
		return status.Error(codes.Unauthenticated, "missing or invalid rpc caller principal")
	}
	return nil
}
