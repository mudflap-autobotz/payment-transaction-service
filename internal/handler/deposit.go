package handler

import (
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/dto"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type DepositHandler struct {
	depositService domain.DepositService
	validate       *validator.Validate
	tracer         trace.Tracer
}

func NewDepositHandler(s domain.DepositService, v *validator.Validate) *DepositHandler {
	return &DepositHandler{
		depositService: s,
		validate:       v,
		tracer:         otel.Tracer("handler.deposit"),
	}
}

// @Summary Initiate a deposit and get a PromptPay QR
// @Tags merchants/deposits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.InitiateDepositRequest true "Deposit payload"
// @Success 201 {object} response.DocResponse[dto.DepositResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 409 {object} response.DocConflictResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/deposits/initiate [post]
func (h *DepositHandler) Initiate(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "DepositHandler.Initiate")
	defer span.End()

	var req dto.InitiateDepositRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.NewValidateFormError(err)
	}

	if err := h.validate.Struct(req); err != nil {
		return response.NewValidateFormError(err)
	}

	actor := actorFrom(c)

	deposit, err := h.depositService.Initiate(ctx, actor, domain.InitiateDeposit{
		MerchantID:        middleware.MerchantIDFromContext(c),
		Amount:            req.Amount,
		MerchantReference: req.MerchantReference,
	})
	if err != nil {
		return err
	}

	return response.Created(c, dto.ToDepositResponse(deposit))
}

// @Summary List deposits of the authenticated merchant
// @Tags merchants/deposits
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param status query string false "Transaction status" Enums(PENDING, SUBMITTED, COMPLETED, FAILED, EXPIRED, CANCELLED)
// @Param sort_by query string false "Sort column" Enums(created_at, amount, status)
// @Param order_by query string false "Sort direction" Enums(asc, desc)
// @Success 200 {object} response.DocPaginateResponse[[]dto.DepositListItemResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/deposits [get]
func (h *DepositHandler) List(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "DepositHandler.List")
	defer span.End()

	query, err := h.queryFrom(c)
	if err != nil {
		return err
	}

	deposits, totalItem, err := h.depositService.List(ctx, query)
	if err != nil {
		return err
	}

	return response.SuccessPagination(c, dto.ToDepositListItemResponses(deposits), response.NewPagination(query.Page, query.Size, totalItem))
}

// @Summary Get a deposit by transaction id
// @Tags merchants/deposits
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction id"
// @Success 200 {object} response.DocResponse[dto.DepositResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 404 {object} response.DocNotFoundResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/deposits/{id} [get]
func (h *DepositHandler) Get(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "DepositHandler.Get")
	defer span.End()

	id, err := idFromPath(c, h.validate)
	if err != nil {
		return err
	}

	deposit, err := h.depositService.Get(ctx, middleware.MerchantIDFromContext(c), id)
	if err != nil {
		return err
	}

	return response.Success(c, dto.ToDepositResponse(deposit))
}

func (h *DepositHandler) queryFrom(c fiber.Ctx) (domain.DepositQuery, error) {
	var query dto.DepositQuery
	if err := c.Bind().Query(&query); err != nil {
		return domain.DepositQuery{}, response.NewValidateFormError(err)
	}

	query.ApplyDefaults()

	if err := h.validate.Struct(query); err != nil {
		return domain.DepositQuery{}, response.NewValidateFormError(err)
	}

	return domain.DepositQuery{
		ListQuery:  query.ToListQuery(),
		MerchantID: middleware.MerchantIDFromContext(c),
		Status:     query.Status,
	}, nil
}
