package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"github.com/mudflap-autobotz/payment-common/logger"
)

func NewRecoveryMiddleware() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Ctx(c.Context()).Error().
					Str("panic", fmt.Sprint(recovered)).
					Str("stack", string(debug.Stack())).
					Msg("recovered from panic")

				err = fmt.Errorf("panic: %v", recovered)
			}
		}()

		return c.Next()
	}
}
