package group

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parseGroupID(value, name string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 || strconv.FormatUint(id, 10) != value {
		return 0, status.Errorf(codes.InvalidArgument, "%s must be a positive integer", name)
	}
	return id, nil
}

func parseGroupIDs(values []string, name string) ([]uint64, error) {
	if len(values) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%ss are required", name)
	}
	ids := make([]uint64, 0, len(values))
	for _, value := range values {
		id, err := parseGroupID(value, name)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
