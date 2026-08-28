package middleware

import (
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func NewCorsMiddleware(cfg *config.Config) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.App.Cors.AllowOrigins,
		AllowMethods:     cfg.App.Cors.AllowMethods,
		AllowHeaders:     cfg.App.Cors.AllowHeaders,
		AllowCredentials: cfg.App.Cors.AllowCredentials,
	})
}
