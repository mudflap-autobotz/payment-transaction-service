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

type WithdrawalHandler struct {
	withdrawalService domain.WithdrawalService
	validate          *validator.Validate
	tracer            trace.Tracer
}

func NewWithdrawalHandler(s domain.WithdrawalService, v *validator.Validate) *WithdrawalHandler {
	return &WithdrawalHandler{
		withdrawalService: s,
		validate:          v,
		tracer:            otel.Tracer("handler.withdrawal"),
	}
}

// @Summary Initiate a withdrawal to a bank account
// @Tags merchants/withdrawals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.InitiateWithdrawalRequest true "Withdrawal payload"
// @Success 201 {object} response.DocResponse[dto.WithdrawalResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 404 {object} response.DocNotFoundResponse
// @Failure 409 {object} response.DocConflictResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/withdrawals/initiate [post]
func (h *WithdrawalHandler) Initiate(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "WithdrawalHandler.Initiate")
	defer span.End()

	var req dto.InitiateWithdrawalRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.NewValidateFormError(err)
	}

	if err := h.validate.Struct(req); err != nil {
		return response.NewValidateFormError(err)
	}

	actor := actorFrom(c)

	withdrawal, err := h.withdrawalService.Initiate(ctx, actor, domain.InitiateWithdrawal{
		MerchantID:        middleware.MerchantIDFromContext(c),
		Amount:            req.Amount,
		MerchantReference: req.MerchantReference,
		BankCode:          req.BankCode,
		AccountNumber:     req.AccountNumber,
		AccountName:       req.AccountName,
	})
	if err != nil {
		return err
	}

	return response.Created(c, dto.ToWithdrawalResponse(withdrawal))
}

// @Summary List withdrawals of the authenticated merchant
// @Tags merchants/withdrawals
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param status query string false "Transaction status" Enums(PENDING, SUBMITTED, COMPLETED, FAILED, EXPIRED, CANCELLED)
// @Param sort_by query string false "Sort column" Enums(created_at, amount, status)
// @Param order_by query string false "Sort direction" Enums(asc, desc)
// @Success 200 {object} response.DocPaginateResponse[[]dto.WithdrawalListItemResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/withdrawals [get]
func (h *WithdrawalHandler) List(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "WithdrawalHandler.List")
	defer span.End()

	query, err := h.queryFrom(c)
	if err != nil {
		return err
	}

	withdrawals, totalItem, err := h.withdrawalService.List(ctx, query)
	if err != nil {
		return err
	}

	return response.SuccessPagination(c, dto.ToWithdrawalListItemResponses(withdrawals), response.NewPagination(query.Page, query.Size, totalItem))
}

// @Summary Get a withdrawal by transaction id
// @Tags merchants/withdrawals
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction id"
// @Success 200 {object} response.DocResponse[dto.WithdrawalResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 404 {object} response.DocNotFoundResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /merchants/withdrawals/{id} [get]
func (h *WithdrawalHandler) Get(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "WithdrawalHandler.Get")
	defer span.End()

	id, err := idFromPath(c, h.validate)
	if err != nil {
		return err
	}

	withdrawal, err := h.withdrawalService.Get(ctx, middleware.MerchantIDFromContext(c), id)
	if err != nil {
		return err
	}

	return response.Success(c, dto.ToWithdrawalResponse(withdrawal))
}

func (h *WithdrawalHandler) queryFrom(c fiber.Ctx) (domain.WithdrawalQuery, error) {
	var query dto.WithdrawalQuery
	if err := c.Bind().Query(&query); err != nil {
		return domain.WithdrawalQuery{}, response.NewValidateFormError(err)
	}

	query.ApplyDefaults()

	if err := h.validate.Struct(query); err != nil {
		return domain.WithdrawalQuery{}, response.NewValidateFormError(err)
	}

	return domain.WithdrawalQuery{
		ListQuery:  query.ToListQuery(),
		MerchantID: middleware.MerchantIDFromContext(c),
		Status:     query.Status,
	}, nil
}
