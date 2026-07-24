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
	id, err := strconv.ParseUint(uid, 10, 64)
	if err != nil || id == 0 || strconv.FormatUint(id, 10) != uid {
		return "", status.Error(codes.Unauthenticated, "missing or invalid user identity")
	}
	return uid, nil
}
