package middleware

import (
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/google/uuid"
)

func NewRatelimitMiddleware(cfg *config.Config) fiber.Handler {

	return limiter.New(limiter.Config{
		Max:        cfg.App.RateLimit.MaxRequests,
		Expiration: cfg.App.RateLimit.Expiration,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too Many Requests",
			})
		},
		Next: func(c fiber.Ctx) bool {
			if c.Path() == "/" || c.Path() == "/healthcheck" {
				return true
			}
			return false
		},
	})
}

func NewMerchantRatelimitMiddleware(cfg *config.Config) fiber.Handler {

	return limiter.New(limiter.Config{
		Max:        cfg.App.IdentityRateLimit.MaxRequests,
		Expiration: cfg.App.IdentityRateLimit.Expiration,
		KeyGenerator: func(c fiber.Ctx) string {
			return MerchantIDFromContext(c).String()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too Many Requests",
			})
		},
		Next: func(c fiber.Ctx) bool {
			return MerchantIDFromContext(c) == uuid.Nil
		},
	})
}
