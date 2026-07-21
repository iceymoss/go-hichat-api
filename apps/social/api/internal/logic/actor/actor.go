package actor

import (
	"context"
	"strconv"

	"github.com/iceymoss/go-hichat-api/pkg/ctxdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UID(ctx context.Context) (string, error) {
	uid := ctxdata.GetUId(ctx)
	if id, err := strconv.ParseUint(uid, 10, 64); err != nil || id == 0 {
		return "", status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	return uid, nil
}
