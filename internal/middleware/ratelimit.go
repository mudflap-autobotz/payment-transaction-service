package middleware

import (
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func NewRatelimitMiddleware(cfg *config.Config) fiber.Handler {

	return limiter.New(limiter.Config{
		Max:        cfg.App.RateLimit.MaxRequests,
		Expiration: cfg.App.RateLimit.Expiration,
		KeyGenerator: func(c fiber.Ctx) string {
			userID := c.Get("X-User-ID")
			if userID != "" && userID != "anonymous" {
				return userID
			}
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
