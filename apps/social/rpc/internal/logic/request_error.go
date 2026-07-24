package logic

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func alreadyHandledError(message, currentResult string) error {
	s := status.New(codes.FailedPrecondition, message)
	withDetails, err := s.WithDetails(&errdetails.ErrorInfo{Reason: "already_handled", Domain: "social.request", Metadata: map[string]string{"current_result": currentResult}})
	if err != nil {
		return s.Err()
	}
	return withDetails.Err()
}
