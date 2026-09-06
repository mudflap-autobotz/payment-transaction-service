package handler

import (
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/dto"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type SummaryHandler struct {
	summaryService domain.SummaryService
	tracer         trace.Tracer
}

func NewSummaryHandler(s domain.SummaryService) *SummaryHandler {
	return &SummaryHandler{
		summaryService: s,
		tracer:         otel.Tracer("handler.summary"),
	}
}

// @Summary Get the authenticated merchant dashboard summary
// @Tags merchants/summary
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.DocResponse[dto.MerchantSummaryResponse]
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/summary [get]
func (h *SummaryHandler) Merchant(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "SummaryHandler.Merchant")
	defer span.End()

	summary, err := h.summaryService.Merchant(ctx, middleware.MerchantIDFromContext(c))
	if err != nil {
		return err
	}

	return response.Success(c, dto.ToMerchantSummaryResponse(summary))
}
