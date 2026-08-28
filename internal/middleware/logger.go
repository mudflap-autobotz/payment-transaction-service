package middleware

import (
	"time"

	"github.com/mudflap-autobotz/payment-common/apperror"
	"github.com/mudflap-autobotz/payment-common/logger"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewLoggerMiddleware(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		startTime := time.Now()

		err := c.Next()

		latency := time.Since(startTime)
		statusCode := statusCodeFromResult(c, err)

		span := trace.SpanFromContext(c.Context())
		spanContext := span.SpanContext()

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else if statusCode >= fiber.StatusBadRequest {
			span.SetStatus(codes.Error, "HTTP status error")
		} else {
			span.SetStatus(codes.Ok, "")
		}

		requestID := requestIDFromHeader(c, spanContext)
		userID := userIDFromHeader(c)

		event := logEventForStatus(logger.Ctx(c.Context()), statusCode)
		event.
			Str("request_id", requestID).
			Str("user_id", userID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Int("status", statusCode).
			Dur("latency", latency)

		if err != nil {
			event = event.Err(err)
		}

		event.Msg(c.Path())

		return err
	}
}

func statusCodeFromResult(c fiber.Ctx, err error) int {
	if err != nil {
		if apiErr, ok := apperror.As(err); ok {
			return apiErr.Code
		}
		if fiberErr, ok := err.(*fiber.Error); ok {
			return fiberErr.Code
		}
		if code, ok := apperror.StatusCode(err); ok {
			return code
		}
		return fiber.StatusInternalServerError
	}
	return c.Response().StatusCode()
}

func requestIDFromHeader(c fiber.Ctx, spanContext trace.SpanContext) string {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		if spanContext.HasTraceID() {
			requestID = spanContext.TraceID().String()
		} else {
			requestID = uuid.New().String()
		}
	}
	c.Set("X-Request-ID", requestID)
	return requestID
}

func userIDFromHeader(c fiber.Ctx) string {
	userID := c.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}
	return userID
}

func logEventForStatus(l *zerolog.Logger, statusCode int) *zerolog.Event {
	switch {
	case statusCode >= fiber.StatusInternalServerError:
		return l.Error()
	case statusCode >= fiber.StatusBadRequest:
		return l.Warn()
	default:
		return l.Info()
	}
}
