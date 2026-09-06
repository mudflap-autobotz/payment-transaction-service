package router

import (
	"github.com/mudflap-autobotz/payment-transaction-service/internal/handler"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func InitMerchantRouter(app *fiber.App, h *handler.Handlers, m *middleware.Middlewares) {
	merchants := app.Group("/api/v1/merchants", m.AuthMerchant, m.MerchantRatelimit)

	merchants.Get("/summary", h.SummaryHandler.Merchant)

	deposits := merchants.Group("/deposits")
	deposits.Post("/initiate", h.DepositHandler.Initiate)
	deposits.Get("", h.DepositHandler.List)
	deposits.Get("/:id", h.DepositHandler.Get)

	withdrawals := merchants.Group("/withdrawals")
	withdrawals.Post("/initiate", h.WithdrawalHandler.Initiate)
	withdrawals.Get("", h.WithdrawalHandler.List)
	withdrawals.Get("/:id", h.WithdrawalHandler.Get)
}
