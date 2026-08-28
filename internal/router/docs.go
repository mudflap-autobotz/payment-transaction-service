package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mudflap-autobotz/payment-transaction-service/docs"
	"github.com/yokeTH/gofiber-scalar/scalar/v3"
)

func InitDocsRouter(app *fiber.App) {
	app.Get("/docs/*", scalar.New(scalar.Config{
		Path:              "/docs",
		Title:             "Payment Transaction Service API",
		FileContentString: string(docs.SwaggerJSON),
	}))
}
