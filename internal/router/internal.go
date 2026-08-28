package router

import (
	"github.com/mudflap-autobotz/payment-service-go-template/internal/handler"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func InitInternalRouter(app *fiber.App, h *handler.Handlers, m *middleware.Middlewares) {
	api := app.Group("/api/v1/internal", m.Auth)

	tigerbaboon := api.Group("/tigerbaboons")
	tigerbaboon.Get("/", h.TigerbaboonHandler.GetList)
	tigerbaboon.Post("/", h.TigerbaboonHandler.Create)
	tigerbaboon.Get("/:id", h.TigerbaboonHandler.GetByID)
	tigerbaboon.Put("/:id", h.TigerbaboonHandler.Update)
	tigerbaboon.Delete("/:id", h.TigerbaboonHandler.Delete)
}
