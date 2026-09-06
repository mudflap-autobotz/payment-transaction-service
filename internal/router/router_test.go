package router_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-common/jwt/jwttest"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain/mocks"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/handler"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/router"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/token"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var routerTestKeys = jwttest.MustGenerateKeys()

type registeredRoute struct {
	method string
	path   string
}

func testConfig() *config.Config {
	cfg := &config.Config{}
	cfg.App.Name = "payment-transaction-service-test"
	cfg.App.RateLimit = config.RateLimitConfig{MaxRequests: 1000, Expiration: time.Minute}
	cfg.App.Cors = config.CorsConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{fiber.MethodGet, fiber.MethodPost},
		AllowHeaders: []string{"*"},
	}
	cfg.JWT = config.JWTConfig{PublicKey: routerTestKeys.PublicPEM}

	return cfg
}

func testHandlers(t *testing.T) *handler.Handlers {
	t.Helper()

	v := handler.NewValidator()

	return &handler.Handlers{
		DepositHandler:    handler.NewDepositHandler(mocks.NewDepositService(t), v),
		WithdrawalHandler: handler.NewWithdrawalHandler(mocks.NewWithdrawalService(t), v),
		SummaryHandler:    handler.NewSummaryHandler(mocks.NewSummaryService(t)),
	}
}

func testMiddlewares(t *testing.T, cfg *config.Config) *middleware.Middlewares {
	t.Helper()

	verifier, err := token.NewVerifier(cfg)
	require.NoError(t, err)

	return middleware.NewMiddlewares(cfg, verifier)
}

func newTestRouter(t *testing.T) *fiber.App {
	t.Helper()

	return newTestRouterWithConfig(t, testConfig())
}

func newTestRouterWithConfig(t *testing.T, cfg *config.Config) *fiber.App {
	t.Helper()

	return router.NewRouter(cfg, testHandlers(t), testMiddlewares(t, cfg))
}

func registeredRoutes(app *fiber.App) map[registeredRoute]bool {
	routes := make(map[registeredRoute]bool)

	for _, route := range app.GetRoutes(true) {
		routes[registeredRoute{method: route.Method, path: route.Path}] = true
	}

	return routes
}

func merchantRoutes() []registeredRoute {
	return []registeredRoute{
		{method: http.MethodGet, path: "/api/v1/merchants/summary"},
		{method: http.MethodPost, path: "/api/v1/merchants/deposits/initiate"},
		{method: http.MethodGet, path: "/api/v1/merchants/deposits"},
		{method: http.MethodGet, path: "/api/v1/merchants/deposits/:id"},
		{method: http.MethodPost, path: "/api/v1/merchants/withdrawals/initiate"},
		{method: http.MethodGet, path: "/api/v1/merchants/withdrawals"},
		{method: http.MethodGet, path: "/api/v1/merchants/withdrawals/:id"},
	}
}

func TestNewRouterServesTheHealthcheck(t *testing.T) {
	app := newTestRouter(t)

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/healthcheck", nil))
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestNewRouterServesTheDocs(t *testing.T) {
	app := newTestRouter(t)

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/docs/", nil))
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestInitMerchantRouterRegistersEveryRoute(t *testing.T) {
	routes := registeredRoutes(newTestRouter(t))

	for _, route := range merchantRoutes() {
		assert.True(t, routes[route], "%s %s is not registered", route.method, route.path)
	}
}

func TestMerchantRoutesRejectAnUnauthenticatedCaller(t *testing.T) {
	app := newTestRouter(t)

	protected := []registeredRoute{
		{method: http.MethodGet, path: "/api/v1/merchants/summary"},
		{method: http.MethodPost, path: "/api/v1/merchants/deposits/initiate"},
		{method: http.MethodGet, path: "/api/v1/merchants/deposits"},
		{method: http.MethodGet, path: "/api/v1/merchants/deposits/0199ae4c-1c8e-7000-8000-000000000001"},
		{method: http.MethodPost, path: "/api/v1/merchants/withdrawals/initiate"},
		{method: http.MethodGet, path: "/api/v1/merchants/withdrawals"},
		{method: http.MethodGet, path: "/api/v1/merchants/withdrawals/0199ae4c-1c8e-7000-8000-000000000001"},
	}

	for _, route := range protected {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			res, err := app.Test(httptest.NewRequest(route.method, route.path, nil))
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		})
	}
}

func TestNewRouterRegistersNoUnexpectedAPIRoutes(t *testing.T) {
	expected := make(map[registeredRoute]bool, len(merchantRoutes()))
	for _, route := range merchantRoutes() {
		expected[route] = true
	}

	for route := range registeredRoutes(newTestRouter(t)) {
		if route.method == http.MethodHead || !isAPIRoute(route.path) {
			continue
		}

		assert.True(t, expected[route], "unexpected route registered: %s %s", route.method, route.path)
	}
}

func isAPIRoute(path string) bool {
	const apiPrefix = "/api/v1/"

	return len(path) >= len(apiPrefix) && path[:len(apiPrefix)] == apiPrefix
}

func newEchoIPRouter(t *testing.T, trustProxy config.TrustProxyConfig) *fiber.App {
	t.Helper()

	cfg := testConfig()
	cfg.App.TrustProxy = trustProxy

	app := newTestRouterWithConfig(t, cfg)
	app.Get("/client-ip", func(c fiber.Ctx) error {
		return c.SendString(c.IP())
	})

	return app
}

func clientIPFor(t *testing.T, app *fiber.App, forwardedFor string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/client-ip", nil)
	if forwardedFor != "" {
		req.Header.Set(fiber.HeaderXForwardedFor, forwardedFor)
	}

	res, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return string(body)
}

func TestClientIPIgnoresForwardedHeaderWhenTrustProxyDisabled(t *testing.T) {
	app := newEchoIPRouter(t, config.TrustProxyConfig{Enabled: false})

	assert.Equal(t, "0.0.0.0", clientIPFor(t, app, "203.0.113.9"))
}

func TestClientIPUsesForwardedHeaderFromATrustedProxy(t *testing.T) {
	app := newEchoIPRouter(t, config.TrustProxyConfig{
		Enabled: true,
		Header:  fiber.HeaderXForwardedFor,
		Proxies: []string{"0.0.0.0"},
	})

	assert.Equal(t, "203.0.113.9", clientIPFor(t, app, "203.0.113.9"))
}

func TestClientIPKeepsTheLeftmostUntrustedHopInAForwardedChain(t *testing.T) {
	app := newEchoIPRouter(t, config.TrustProxyConfig{
		Enabled: true,
		Header:  fiber.HeaderXForwardedFor,
		Private: true,
		Proxies: []string{"0.0.0.0"},
	})

	assert.Equal(t, "203.0.113.9", clientIPFor(t, app, "203.0.113.9, 10.0.0.1, 10.0.0.2"))
}

func TestClientIPRejectsAMalformedForwardedHeader(t *testing.T) {
	app := newEchoIPRouter(t, config.TrustProxyConfig{
		Enabled: true,
		Header:  fiber.HeaderXForwardedFor,
		Proxies: []string{"0.0.0.0"},
	})

	assert.Equal(t, "0.0.0.0", clientIPFor(t, app, "not-an-ip"))
}
