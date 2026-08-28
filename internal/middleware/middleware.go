package middleware

import (
	"github.com/mudflap-autobotz/payment-service-go-template/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

var MiddlewareSet = wire.NewSet(
	NewMiddlewares,
)

type Middlewares struct {
	Auth      fiber.Handler
	Cors      fiber.Handler
	Logger    fiber.Handler
	Ratelimit fiber.Handler
	Recovery  fiber.Handler
}

func NewMiddlewares(cfg *config.Config) *Middlewares {
	return &Middlewares{
		Auth:      NewAuthMiddleware(),
		Cors:      NewCorsMiddleware(cfg),
		Logger:    NewLoggerMiddleware(cfg),
		Ratelimit: NewRatelimitMiddleware(cfg),
		Recovery:  NewRecoveryMiddleware(),
	}
}
