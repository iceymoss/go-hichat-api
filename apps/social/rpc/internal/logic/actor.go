package logic

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateScopedActor(actorUID string, legacyUIDs ...string) (string, error) {
	if actorUID == "" {
		return "", status.Error(codes.Unauthenticated, "actor uid is required")
	}
	if id, err := strconv.ParseUint(actorUID, 10, 64); err != nil || id == 0 {
		return "", status.Error(codes.InvalidArgument, "actor uid must be a positive integer")
	}
	for _, legacyUID := range legacyUIDs {
		if legacyUID != "" && legacyUID != actorUID {
			return "", status.Error(codes.PermissionDenied, "actor uid does not match legacy user id")
		}
	}
	return actorUID, nil
}
