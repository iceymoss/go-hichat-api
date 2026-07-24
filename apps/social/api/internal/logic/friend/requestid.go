package friend

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parseRequestID(value string, legacy int32) (uint64, error) {
	if value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil || id == 0 {
			return 0, status.Error(codes.InvalidArgument, "friend request id must be a positive integer")
		}
		return id, nil
	}
	if legacy <= 0 {
		return 0, status.Error(codes.InvalidArgument, "friend request id must be positive")
	}
	return uint64(legacy), nil
}

func parseRequestIDs(values []string) ([]uint64, error) {
	ids := make([]uint64, 0, len(values))
	for _, value := range values {
		id, err := parseRequestID(value, 0)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
