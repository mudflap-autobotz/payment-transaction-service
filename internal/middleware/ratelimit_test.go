package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ratelimitPath   = "/rate-limited"
	healthcheckPath = "/healthcheck"
	userIDHeader    = "X-User-ID"
)

func ratelimitConfig(maxRequests int) *config.Config {
	cfg := &config.Config{}
	cfg.App.RateLimit = config.RateLimitConfig{MaxRequests: maxRequests, Expiration: time.Minute}
	cfg.App.IdentityRateLimit = config.RateLimitConfig{MaxRequests: maxRequests, Expiration: time.Minute}

	return cfg
}

func newRatelimitApp(limit fiber.Handler, paths ...string) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: response.Error})
	app.Use(limit)

	for _, path := range paths {
		app.Get(path, okHandler)
	}

	return app
}

func newIdentityRatelimitApp(limit fiber.Handler, subject fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: response.Error})
	app.Get(ratelimitPath, subject, limit, okHandler)

	return app
}

func okHandler(c fiber.Ctx) error {
	return response.Success(c, fiber.Map{"status": "ok"})
}

func authenticatedAs(t *testing.T, merchantID uuid.UUID) fiber.Handler {
	t.Helper()

	auth := middleware.NewAuthMerchantMiddleware(newVerifier(t))

	bearer := "Bearer " + signedToken(t, testKeys, time.Hour, commonjwt.Claims{
		UserID: merchantID,
		Email:  merchantEmail,
		Type:   domain.TokenTypeMerchant,
	})

	return func(c fiber.Ctx) error {
		c.Request().Header.Set(fiber.HeaderAuthorization, bearer)

		return auth(c)
	}
}

func statusesFor(t *testing.T, app *fiber.App, path string, attempts int, headers map[string]string) []int {
	t.Helper()

	statuses := make([]int, 0, attempts)

	for range attempts {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		for name, value := range headers {
			request.Header.Set(name, value)
		}

		res, err := app.Test(request)
		require.NoError(t, err)

		statuses = append(statuses, res.StatusCode)

		_ = res.Body.Close()
	}

	return statuses
}

func TestNewRatelimitMiddleware(t *testing.T) {
	t.Run("rejects a caller that exceeds the quota", func(t *testing.T) {
		app := newRatelimitApp(middleware.NewRatelimitMiddleware(ratelimitConfig(2)), ratelimitPath)

		statuses := statusesFor(t, app, ratelimitPath, 3, nil)

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}, statuses)
	})

	t.Run("leaves the healthcheck uncounted", func(t *testing.T) {
		app := newRatelimitApp(middleware.NewRatelimitMiddleware(ratelimitConfig(2)), healthcheckPath)

		statuses := statusesFor(t, app, healthcheckPath, 5, nil)

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK}, statuses)
	})

	t.Run("keys on the address and ignores a spoofed user header", func(t *testing.T) {
		app := newRatelimitApp(middleware.NewRatelimitMiddleware(ratelimitConfig(2)), ratelimitPath)

		spoofed := map[string]string{userIDHeader: uuid.NewString()}

		statuses := statusesFor(t, app, ratelimitPath, 3, spoofed)

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}, statuses)
	})

	t.Run("keeps one bucket for a caller that rotates the user header", func(t *testing.T) {
		app := newRatelimitApp(middleware.NewRatelimitMiddleware(ratelimitConfig(2)), ratelimitPath)

		statuses := make([]int, 0, 3)
		for range 3 {
			rotated := map[string]string{userIDHeader: uuid.NewString()}
			statuses = append(statuses, statusesFor(t, app, ratelimitPath, 1, rotated)...)
		}

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}, statuses)
	})
}

func TestNewMerchantRatelimitMiddleware(t *testing.T) {
	t.Run("rejects a merchant that exceeds the quota", func(t *testing.T) {
		app := newIdentityRatelimitApp(
			middleware.NewMerchantRatelimitMiddleware(ratelimitConfig(2)),
			authenticatedAs(t, uuid.New()),
		)

		statuses := statusesFor(t, app, ratelimitPath, 3, nil)

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}, statuses)
	})

	t.Run("counts two merchants from one address separately", func(t *testing.T) {
		limit := middleware.NewMerchantRatelimitMiddleware(ratelimitConfig(2))

		first := newIdentityRatelimitApp(limit, authenticatedAs(t, uuid.New()))
		second := newIdentityRatelimitApp(limit, authenticatedAs(t, uuid.New()))

		assert.Equal(t, []int{http.StatusOK, http.StatusOK}, statusesFor(t, first, ratelimitPath, 2, nil))
		assert.Equal(t, []int{http.StatusOK, http.StatusOK}, statusesFor(t, second, ratelimitPath, 2, nil))
	})

	t.Run("skips a request that carries no identity", func(t *testing.T) {
		app := newIdentityRatelimitApp(
			middleware.NewMerchantRatelimitMiddleware(ratelimitConfig(2)),
			func(c fiber.Ctx) error { return c.Next() },
		)

		statuses := statusesFor(t, app, ratelimitPath, 5, nil)

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK, http.StatusOK}, statuses)
	})

	t.Run("ignores a spoofed user header", func(t *testing.T) {
		app := newIdentityRatelimitApp(
			middleware.NewMerchantRatelimitMiddleware(ratelimitConfig(2)),
			authenticatedAs(t, uuid.New()),
		)

		statuses := make([]int, 0, 3)
		for range 3 {
			rotated := map[string]string{userIDHeader: uuid.NewString()}
			statuses = append(statuses, statusesFor(t, app, ratelimitPath, 1, rotated)...)
		}

		assert.Equal(t, []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}, statuses)
	})
}
