package grpcapi

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
)

// toStatus maps domain errors to gRPC status errors, hiding internal details.
func toStatus(err error) error {
	if err == nil {
		return nil
	}

	var notFound *domainerrors.NotFoundError
	var validation *domainerrors.ValidationError
	var forbidden *domainerrors.ForbiddenError
	var isDir *domainerrors.IsDirectoryError
	var notDir *domainerrors.NotDirectoryError

	switch {
	case errors.Is(err, domainerrors.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domainerrors.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domainerrors.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &notFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &validation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &isDir):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.As(err, &notDir):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &forbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
