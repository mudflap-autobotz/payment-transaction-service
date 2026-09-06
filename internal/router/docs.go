package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mudflap-autobotz/payment-transaction-service/docs"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/yokeTH/gofiber-scalar/scalar/v3"
)

func InitDocsRouter(app *fiber.App, cfg *config.Config) {
	app.Get("/docs/*", scalar.New(scalar.Config{
		Path:              "/docs",
		Title:             cfg.App.Name,
		FileContentString: string(docs.SwaggerJSON),
	}))
}
