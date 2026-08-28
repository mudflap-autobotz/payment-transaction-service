package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
)

var HandlerSet = wire.NewSet(
	NewValidator,
	NewTigerbaboonHandler,

	wire.Struct(new(Handlers), "*"),
)

func NewValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

type Handlers struct {
	TigerbaboonHandler *TigerbaboonHandler
}
