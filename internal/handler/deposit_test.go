package handler_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain/mocks"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/dto"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/handler"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const depositsPath = "/api/v1/merchants/deposits"

func setupDepositApp(t *testing.T) (*mocks.DepositService, *fiber.App) {
	t.Helper()

	svc := mocks.NewDepositService(t)
	h := handler.NewDepositHandler(svc, handler.NewValidator())

	app := newApp()
	deposits := app.Group(depositsPath, middleware.NewAuthMerchantMiddleware(testVerifier(t)))
	deposits.Post("/initiate", h.Initiate)
	deposits.Get("", h.List)
	deposits.Get("/:id", h.Get)

	return svc, app
}

func depositFixture(merchantID uuid.UUID) *domain.Deposit {
	return &domain.Deposit{
		Transaction: domain.Transaction{
			ID:                uuid.New(),
			MerchantID:        merchantID,
			Type:              domain.TransactionTypeDeposit,
			Amount:            "500.00",
			Status:            domain.TransactionStatusPending,
			MerchantReference: "ORDER-1",
			CreatedAt:         time.Now().UTC().Truncate(time.Second),
			UpdatedAt:         time.Now().UTC().Truncate(time.Second),
		},
		QRCode:     "00020101021229370016A000000677010111",
		QRType:     domain.QRTypePromptPay,
		QRExpiryAt: time.Now().UTC().Add(30 * time.Minute).Truncate(time.Second),
	}
}

func TestDepositHandlerInitiate(t *testing.T) {
	t.Run("returns 201 with the qr payload", func(t *testing.T) {
		svc, app := setupDepositApp(t)
		merchantID := uuid.New()
		created := depositFixture(merchantID)

		svc.EXPECT().
			Initiate(mock.Anything, mock.MatchedBy(func(actor domain.Actor) bool {
				return actor.Email == merchantEmail && actor.UserID == uuid.Nil
			}), domain.InitiateDeposit{
				MerchantID:        merchantID,
				Amount:            "500.00",
				MerchantReference: "ORDER-1",
			}).
			Return(created, nil).
			Once()

		res := doAuthorizedRequest(t, app, http.MethodPost, depositsPath+"/initiate", merchantBearer(t, merchantID),
			dto.InitiateDepositRequest{Amount: "500.00", MerchantReference: "ORDER-1"})

		require.Equal(t, http.StatusCreated, res.StatusCode)

		body := decodeBody[dataEnvelope[dto.DepositResponse]](t, res)
		assert.Equal(t, created.ID.String(), body.Data.TransactionID)
		assert.Equal(t, merchantID.String(), body.Data.MerchantID)
		assert.Equal(t, "500.00", body.Data.Amount)
		assert.Equal(t, created.QRCode, body.Data.QRCode)
		assert.Equal(t, domain.QRTypePromptPay, body.Data.QRType)
		assert.False(t, body.Data.Expired)
	})

	t.Run("returns 401 without a bearer token", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		res := doRequest(t, app, http.MethodPost, depositsPath+"/initiate",
			dto.InitiateDepositRequest{Amount: "500.00"})

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		svc.AssertNotCalled(t, "Initiate")
	})

	t.Run("returns 401 for an admin token", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		res := doAuthorizedRequest(t, app, http.MethodPost, depositsPath+"/initiate",
			bearerFor(t, uuid.New(), "admin@example.com", domain.TokenTypeAdmin),
			dto.InitiateDepositRequest{Amount: "500.00"})

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		svc.AssertNotCalled(t, "Initiate")
	})

	t.Run("returns 400 when amount is missing or not numeric", func(t *testing.T) {
		for _, amount := range []string{"", "abc"} {
			svc, app := setupDepositApp(t)

			res := doAuthorizedRequest(t, app, http.MethodPost, depositsPath+"/initiate",
				merchantBearer(t, uuid.New()), dto.InitiateDepositRequest{Amount: amount})

			assert.Equal(t, http.StatusBadRequest, res.StatusCode, amount)
			svc.AssertNotCalled(t, "Initiate")
		}
	})

	t.Run("returns 400 when the service rejects the amount", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		svc.EXPECT().Initiate(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrInvalidAmount).Once()

		res := doAuthorizedRequest(t, app, http.MethodPost, depositsPath+"/initiate",
			merchantBearer(t, uuid.New()), dto.InitiateDepositRequest{Amount: "1.00"})

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("returns 409 for a duplicate merchant reference", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		svc.EXPECT().Initiate(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrDuplicateReference).Once()

		res := doAuthorizedRequest(t, app, http.MethodPost, depositsPath+"/initiate",
			merchantBearer(t, uuid.New()), dto.InitiateDepositRequest{Amount: "500.00", MerchantReference: "ORDER-1"})

		assert.Equal(t, http.StatusConflict, res.StatusCode)
	})

	t.Run("returns 500 on an unexpected failure", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		svc.EXPECT().Initiate(mock.Anything, mock.Anything, mock.Anything).Return(nil, errDatabase).Once()

		res := doAuthorizedRequest(t, app, http.MethodPost, depositsPath+"/initiate",
			merchantBearer(t, uuid.New()), dto.InitiateDepositRequest{Amount: "500.00"})

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})
}

func TestDepositHandlerGet(t *testing.T) {
	t.Run("returns 200 with the deposit", func(t *testing.T) {
		svc, app := setupDepositApp(t)
		merchantID := uuid.New()
		deposit := depositFixture(merchantID)

		svc.EXPECT().Get(mock.Anything, merchantID, deposit.ID).Return(deposit, nil).Once()

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath+"/"+deposit.ID.String(),
			merchantBearer(t, merchantID), nil)

		require.Equal(t, http.StatusOK, res.StatusCode)

		body := decodeBody[dataEnvelope[dto.DepositResponse]](t, res)
		assert.Equal(t, deposit.ID.String(), body.Data.TransactionID)
	})

	t.Run("returns 400 for a non-uuid id", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath+"/"+invalidUUID,
			merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		svc.AssertNotCalled(t, "Get")
	})

	t.Run("returns 404 for another merchant's deposit", func(t *testing.T) {
		svc, app := setupDepositApp(t)
		merchantID := uuid.New()
		id := uuid.New()

		svc.EXPECT().Get(mock.Anything, merchantID, id).Return(nil, domain.ErrTransactionNotFound).Once()

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath+"/"+id.String(),
			merchantBearer(t, merchantID), nil)

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("returns 401 without a bearer token", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		res := doRequest(t, app, http.MethodGet, depositsPath+"/"+uuid.New().String(), nil)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		svc.AssertNotCalled(t, "Get")
	})
}

func TestDepositHandlerList(t *testing.T) {
	t.Run("returns 200 with pagination scoped to the merchant", func(t *testing.T) {
		svc, app := setupDepositApp(t)
		merchantID := uuid.New()
		deposits := []domain.Deposit{*depositFixture(merchantID)}

		svc.EXPECT().
			List(mock.Anything, domain.DepositQuery{
				ListQuery:  domain.ListQuery{Page: 1, Size: 10, SortBy: "created_at", OrderBy: "desc"},
				MerchantID: merchantID,
				Status:     domain.TransactionStatusPending,
			}).
			Return(deposits, int64(1), nil).
			Once()

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath+"?status=PENDING",
			merchantBearer(t, merchantID), nil)

		require.Equal(t, http.StatusOK, res.StatusCode)

		body := decodeBody[paginateEnvelope[[]dto.DepositListItemResponse]](t, res)
		require.Len(t, body.Data, 1)
		assert.Equal(t, deposits[0].ID.String(), body.Data[0].TransactionID)
		assert.Equal(t, int64(1), body.Pagination.TotalItem)
	})

	t.Run("returns 400 for an unknown status", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath+"?status=NOPE",
			merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		svc.AssertNotCalled(t, "List")
	})

	t.Run("returns 400 for a sort column outside the whitelist", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath+"?sort_by=qr_code",
			merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		svc.AssertNotCalled(t, "List")
	})

	t.Run("returns 500 on an unexpected failure", func(t *testing.T) {
		svc, app := setupDepositApp(t)

		svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, int64(0), errDatabase).Once()

		res := doAuthorizedRequest(t, app, http.MethodGet, depositsPath, merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		body := decodeBody[errorEnvelope](t, res)
		assert.NotEmpty(t, body.Message)
	})
}
