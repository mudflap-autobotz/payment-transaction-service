package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain/mocks"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/dto"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/handler"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*mocks.TigerbaboonService, *fiber.App) {
	t.Helper()

	svc := mocks.NewTigerbaboonService(t)
	h := handler.NewTigerbaboonHandler(svc, handler.NewValidator())

	app := fiber.New(fiber.Config{ErrorHandler: response.Error})
	app.Get("/tigerbaboons", h.GetList)
	app.Get("/tigerbaboons/:id", h.GetByID)
	app.Post("/tigerbaboons", h.Create)
	app.Put("/tigerbaboons/:id", h.Update)
	app.Delete("/tigerbaboons/:id", h.Delete)

	return svc, app
}

func doRequest(t *testing.T, app *fiber.App, method, target, body string) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}

func decodeData(t *testing.T, body []byte) dto.TigerbaboonResponse {
	t.Helper()

	var resp response.Response[dto.TigerbaboonResponse]
	require.NoError(t, json.Unmarshal(body, &resp))

	return resp.Data
}

func TestTigerbaboonHandlerGetByID(t *testing.T) {
	t.Run("returns 200 with tigerbaboon", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().GetByID(mock.Anything, 1).Return(&domain.Tigerbaboon{ID: 1, Username: "alice"}, nil)

		resp, body := doRequest(t, app, http.MethodGet, "/tigerbaboons/1", "")

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, dto.TigerbaboonResponse{ID: 1, Username: "alice"}, decodeData(t, body))
	})

	t.Run("returns 400 when id is not an integer", func(t *testing.T) {
		svc, app := setup(t)

		resp, _ := doRequest(t, app, http.MethodGet, "/tigerbaboons/abc", "")

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		svc.AssertNotCalled(t, "GetByID")
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().GetByID(mock.Anything, 99).Return(nil, domain.ErrTigerbaboonNotFound)

		resp, _ := doRequest(t, app, http.MethodGet, "/tigerbaboons/99", "")

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestTigerbaboonHandlerGetList(t *testing.T) {
	t.Run("returns 200 with pagination", func(t *testing.T) {
		svc, app := setup(t)
		expectedQuery := domain.ListQuery{Page: 1, Size: 10, OrderBy: "desc", SortBy: "id"}
		svc.EXPECT().GetList(mock.Anything, expectedQuery).Return(&domain.TigerbaboonList{
			Tigerbaboons: []domain.Tigerbaboon{{ID: 1, Username: "alice"}, {ID: 2, Username: "bob"}},
			Total:        2,
		}, nil)

		resp, body := doRequest(t, app, http.MethodGet, "/tigerbaboons", "")

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var paginated response.ResponsePaginate[[]dto.TigerbaboonResponse]
		require.NoError(t, json.Unmarshal(body, &paginated))
		assert.Len(t, paginated.Data, 2)
	})

	t.Run("returns 400 on invalid sort_by", func(t *testing.T) {
		svc, app := setup(t)

		resp, _ := doRequest(t, app, http.MethodGet, "/tigerbaboons?sort_by=invalid", "")

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		svc.AssertNotCalled(t, "GetList")
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().GetList(mock.Anything, mock.Anything).Return(nil, assert.AnError)

		resp, _ := doRequest(t, app, http.MethodGet, "/tigerbaboons", "")

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestTigerbaboonHandlerCreate(t *testing.T) {
	t.Run("returns 201 with created tigerbaboon", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().Create(mock.Anything, "alice").Return(&domain.Tigerbaboon{ID: 1, Username: "alice"}, nil)

		resp, body := doRequest(t, app, http.MethodPost, "/tigerbaboons", `{"username":"alice"}`)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, dto.TigerbaboonResponse{ID: 1, Username: "alice"}, decodeData(t, body))
	})

	t.Run("returns 400 when username too short", func(t *testing.T) {
		svc, app := setup(t)

		resp, _ := doRequest(t, app, http.MethodPost, "/tigerbaboons", `{"username":"ab"}`)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		svc.AssertNotCalled(t, "Create")
	})

	t.Run("returns 409 on duplicate username", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().Create(mock.Anything, "alice").Return(nil, domain.ErrTigerbaboonConflict)

		resp, _ := doRequest(t, app, http.MethodPost, "/tigerbaboons", `{"username":"alice"}`)

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestTigerbaboonHandlerUpdate(t *testing.T) {
	t.Run("returns 200 with updated tigerbaboon", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().Update(mock.Anything, 1, "bob").Return(&domain.Tigerbaboon{ID: 1, Username: "bob"}, nil)

		resp, body := doRequest(t, app, http.MethodPut, "/tigerbaboons/1", `{"username":"bob"}`)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, dto.TigerbaboonResponse{ID: 1, Username: "bob"}, decodeData(t, body))
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().Update(mock.Anything, 99, "bob").Return(nil, domain.ErrTigerbaboonNotFound)

		resp, _ := doRequest(t, app, http.MethodPut, "/tigerbaboons/99", `{"username":"bob"}`)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestTigerbaboonHandlerDelete(t *testing.T) {
	t.Run("returns 200 on success", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().Delete(mock.Anything, 1).Return(nil)

		resp, _ := doRequest(t, app, http.MethodDelete, "/tigerbaboons/1", "")

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		svc, app := setup(t)
		svc.EXPECT().Delete(mock.Anything, 99).Return(domain.ErrTigerbaboonNotFound)

		resp, _ := doRequest(t, app, http.MethodDelete, "/tigerbaboons/99", "")

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
