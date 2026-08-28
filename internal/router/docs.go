package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mudflap-autobotz/payment-service-go-template/docs"
	"github.com/yokeTH/gofiber-scalar/scalar/v3"
)

func InitDocsRouter(app *fiber.App) {
	app.Get("/docs/*", scalar.New(scalar.Config{
		Path:              "/docs",
		Title:             "Go Template API",
		FileContentString: string(docs.SwaggerJSON),
	}))
}
