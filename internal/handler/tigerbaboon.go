package handler

import (
	"strconv"

	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/dto"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type TigerbaboonHandler struct {
	tigerbaboonService domain.TigerbaboonService
	validate           *validator.Validate
	tracer             trace.Tracer
}

func NewTigerbaboonHandler(s domain.TigerbaboonService, v *validator.Validate) *TigerbaboonHandler {
	return &TigerbaboonHandler{
		tigerbaboonService: s,
		validate:           v,
		tracer:             otel.Tracer("handler.tigerbaboon"),
	}
}

// @Summary Get tigerbaboon by id
// @Tags internal/tigerbaboons
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tigerbaboon id"
// @Success 200 {object} response.DocResponse[dto.TigerbaboonResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 404 {object} response.DocNotFoundResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /internal/tigerbaboons/{id} [get]
func (h *TigerbaboonHandler) GetByID(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "TigerbaboonHandler.GetByID")
	defer span.End()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.NewParameterError("id", "must be an integer")
	}

	tigerbaboon, err := h.tigerbaboonService.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return response.Success(c, toTigerbaboonResponse(*tigerbaboon))
}

// @Summary List tigerbaboons
// @Tags internal/tigerbaboons
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param search query string false "Search by username"
// @Param sort_by query string false "Sort column" Enums(id, username)
// @Param order_by query string false "Sort direction" Enums(asc, desc)
// @Success 200 {object} response.DocPaginateResponse[[]dto.TigerbaboonResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /internal/tigerbaboons [get]
func (h *TigerbaboonHandler) GetList(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "TigerbaboonHandler.GetList")
	defer span.End()

	var req dto.GetTigerbaboonListRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.NewValidateFormError(err)
	}

	req.ApplyDefaults()

	if err := h.validate.Struct(req); err != nil {
		return response.NewValidateFormError(err)
	}

	list, err := h.tigerbaboonService.GetList(ctx, req.ToListQuery())
	if err != nil {
		return err
	}

	tigerbaboons := make([]dto.TigerbaboonResponse, 0, len(list.Tigerbaboons))
	for _, tigerbaboon := range list.Tigerbaboons {
		tigerbaboons = append(tigerbaboons, toTigerbaboonResponse(tigerbaboon))
	}

	return response.SuccessPagination(c, tigerbaboons, response.NewPagination(req.Page, req.Size, list.Total))
}

// @Summary Create a tigerbaboon
// @Tags internal/tigerbaboons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateTigerbaboonRequest true "Tigerbaboon payload"
// @Success 201 {object} response.DocCreatedResponse[dto.TigerbaboonResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 409 {object} response.DocConflictResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /internal/tigerbaboons [post]
func (h *TigerbaboonHandler) Create(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "TigerbaboonHandler.Create")
	defer span.End()

	var req dto.CreateTigerbaboonRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.NewValidateFormError(err)
	}

	if err := h.validate.Struct(req); err != nil {
		return response.NewValidateFormError(err)
	}

	tigerbaboon, err := h.tigerbaboonService.Create(ctx, req.Username)
	if err != nil {
		return err
	}

	return response.Created(c, toTigerbaboonResponse(*tigerbaboon))
}

// @Summary Update a tigerbaboon
// @Tags internal/tigerbaboons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tigerbaboon id"
// @Param body body dto.UpdateTigerbaboonRequest true "Tigerbaboon payload"
// @Success 200 {object} response.DocResponse[dto.TigerbaboonResponse]
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 404 {object} response.DocNotFoundResponse
// @Failure 409 {object} response.DocConflictResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /internal/tigerbaboons/{id} [put]
func (h *TigerbaboonHandler) Update(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "TigerbaboonHandler.Update")
	defer span.End()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.NewParameterError("id", "must be an integer")
	}

	var req dto.UpdateTigerbaboonRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.NewValidateFormError(err)
	}

	if err := h.validate.Struct(req); err != nil {
		return response.NewValidateFormError(err)
	}

	tigerbaboon, err := h.tigerbaboonService.Update(ctx, id, req.Username)
	if err != nil {
		return err
	}

	return response.Success(c, toTigerbaboonResponse(*tigerbaboon))
}

// @Summary Delete a tigerbaboon
// @Tags internal/tigerbaboons
// @Produce json
// @Security BearerAuth
// @Param id path int true "Tigerbaboon id"
// @Success 200 {object} response.DocEmptyResponse
// @Failure 400 {object} response.DocBadRequestResponse
// @Failure 401 {object} response.DocUnauthorizedResponse
// @Failure 404 {object} response.DocNotFoundResponse
// @Failure 500 {object} response.DocInternalServerErrorResponse
// @Router /internal/tigerbaboons/{id} [delete]
func (h *TigerbaboonHandler) Delete(c fiber.Ctx) error {
	ctx, span := h.tracer.Start(c.Context(), "TigerbaboonHandler.Delete")
	defer span.End()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.NewParameterError("id", "must be an integer")
	}

	if err := h.tigerbaboonService.Delete(ctx, id); err != nil {
		return err
	}

	return response.SuccessEmpty(c)
}

func toTigerbaboonResponse(tigerbaboon domain.Tigerbaboon) dto.TigerbaboonResponse {
	return dto.TigerbaboonResponse{
		ID:       tigerbaboon.ID,
		Username: tigerbaboon.Username,
	}
}
