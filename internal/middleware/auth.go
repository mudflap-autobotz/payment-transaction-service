package middleware

import (
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
)

func NewAuthMiddleware() fiber.Handler {
	tracer := otel.Tracer("middleware.auth")

	return func(c fiber.Ctx) error {
		ctx, span := tracer.Start(c.Context(), "AuthMiddleware")
		defer span.End()

		_ = ctx

		return c.Next()
	}
}
