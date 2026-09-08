package middleware

import (
	"github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
)

var MiddlewareSet = wire.NewSet(
	NewRatelimitStorage,
	NewMiddlewares,
)

type Middlewares struct {
	AuthMerchant      fiber.Handler
	Cors              fiber.Handler
	Logger            fiber.Handler
	Ratelimit         fiber.Handler
	MerchantRatelimit fiber.Handler
	Recovery          fiber.Handler
}

func NewMiddlewares(cfg *config.Config, verifier *jwt.Verifier, storage fiber.Storage) *Middlewares {
	return &Middlewares{
		AuthMerchant:      NewAuthMerchantMiddleware(verifier),
		Cors:              NewCorsMiddleware(cfg),
		Logger:            NewLoggerMiddleware(cfg),
		Ratelimit:         NewRatelimitMiddleware(cfg, storage),
		MerchantRatelimit: NewMerchantRatelimitMiddleware(cfg, storage),
		Recovery:          NewRecoveryMiddleware(),
	}
}
