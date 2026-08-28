package domain

import (
	"net/http"

	"github.com/mudflap-autobotz/payment-common/apperror"
)

var (
	ErrNotFound     = apperror.ErrNotFound
	ErrConflict     = apperror.ErrConflict
	ErrInvalidInput = apperror.ErrInvalidInput
	ErrUnauthorized = apperror.ErrUnauthorized
	ErrForbidden    = apperror.ErrForbidden
)

func NotFoundError(entity string) error {
	return apperror.NotFoundError(entity)
}

func ConflictError(entity string) error {
	return apperror.ConflictError(entity)
}

func ConflictReasonError(reason string) error {
	return apperror.New(http.StatusConflict, reason, ErrConflict)
}

func InvalidInputError(reason string) error {
	return apperror.InvalidInputError(reason)
}

func UnauthorizedError(reason string) error {
	return apperror.UnauthorizedError(reason)
}

func ForbiddenError(reason string) error {
	return apperror.ForbiddenError(reason)
}
