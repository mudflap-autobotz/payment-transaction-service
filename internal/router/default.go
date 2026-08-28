package router

import (
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/handler"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"github.com/rs/zerolog/log"
)

var RouterSet = wire.NewSet(NewRouter)

func NewRouter(cfg *config.Config, h *handler.Handlers, m *middleware.Middlewares) *fiber.App {

	log.Info().Msg("Setup router")

	app := fiber.New(newFiberConfig(cfg))

	app.Use(fiberotel.Middleware())
	app.Use(m.Recovery)
	app.Use(m.Cors)
	app.Use(m.Logger)
	app.Use(m.Ratelimit)

	app.Get("/healthcheck", func(c fiber.Ctx) error {
		return response.Success(c, fiber.Map{"status": "ok"})
	})

	InitDocsRouter(app)

	InitMerchantRouter(app, h, m)

	return app
}

func newFiberConfig(cfg *config.Config) fiber.Config {
	fiberCfg := fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: response.Error,
		BodyLimit:    cfg.App.BodyLimit,
	}

	if !cfg.App.TrustProxy.Enabled {
		return fiberCfg
	}

	fiberCfg.ProxyHeader = cfg.App.TrustProxy.Header
	fiberCfg.EnableIPValidation = true
	fiberCfg.TrustProxy = true
	fiberCfg.TrustProxyConfig = fiber.TrustProxyConfig{
		Private:  cfg.App.TrustProxy.Private,
		Loopback: cfg.App.TrustProxy.Loopback,
		Proxies:  cfg.App.TrustProxy.Proxies,
	}

	return fiberCfg
}
