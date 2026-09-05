package middleware

import (
	"strings"

	"github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type localsKey int

const (
	localsKeyMerchantID localsKey = iota
	localsKeyEmail
)

const bearerPrefix = "Bearer "

func NewAuthMerchantMiddleware(verifier *jwt.Verifier) fiber.Handler {
	tracer := otel.Tracer("middleware.auth")

	return func(c fiber.Ctx) error {
		ctx, span := tracer.Start(c.Context(), "AuthMerchantMiddleware")
		defer span.End()

		rawToken, hasPrefix := strings.CutPrefix(c.Get(fiber.HeaderAuthorization), bearerPrefix)
		if !hasPrefix || rawToken == "" {
			return domain.UnauthorizedError("missing bearer token")
		}

		claims, err := verifier.Parse(rawToken)
		if err != nil || claims.Type != domain.TokenTypeMerchant {
			return domain.UnauthorizedError("invalid or expired token")
		}

		fiber.Locals(c, localsKeyMerchantID, claims.UserID)
		fiber.Locals(c, localsKeyEmail, claims.Email)

		c.SetContext(ctx)

		return c.Next()
	}
}

func MerchantIDFromContext(c fiber.Ctx) uuid.UUID {
	return fiber.Locals[uuid.UUID](c, localsKeyMerchantID)
}

func EmailFromContext(c fiber.Ctx) string {
	return fiber.Locals[string](c, localsKeyEmail)
}
