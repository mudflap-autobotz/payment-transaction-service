package domain

import "github.com/mudflap-autobotz/payment-common/apperror"

var (
	ErrNotFound     = apperror.ErrNotFound
	ErrConflict     = apperror.ErrConflict
	ErrInvalidInput = apperror.ErrInvalidInput
	ErrUnauthorized = apperror.ErrUnauthorized
)

func NotFoundError(entity string) error {
	return apperror.NotFoundError(entity)
}

func ConflictError(entity string) error {
	return apperror.ConflictError(entity)
}
