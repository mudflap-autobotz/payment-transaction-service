package handler

import (
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/dto"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/google/wire"
)

var HandlerSet = wire.NewSet(
	NewValidator,
	NewDepositHandler,
	NewWithdrawalHandler,

	wire.Struct(new(Handlers), "*"),
)

func NewValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

type Handlers struct {
	DepositHandler    *DepositHandler
	WithdrawalHandler *WithdrawalHandler
}

func actorFrom(c fiber.Ctx) domain.Actor {
	return domain.Actor{
		Email:     middleware.EmailFromContext(c),
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	}
}

func idFromPath(c fiber.Ctx, v *validator.Validate) (uuid.UUID, error) {
	var param dto.IDParam
	if err := c.Bind().URI(&param); err != nil {
		return uuid.Nil, response.NewParameterError("id", "must be a valid uuid")
	}

	if err := v.Struct(param); err != nil {
		return uuid.Nil, response.NewParameterError("id", "must be a valid uuid")
	}

	id, err := uuid.Parse(param.ID)
	if err != nil {
		return uuid.Nil, response.NewParameterError("id", "must be a valid uuid")
	}

	return id, nil
}
