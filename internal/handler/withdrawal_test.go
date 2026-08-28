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

const withdrawalsPath = "/api/v1/merchants/withdrawals"

func setupWithdrawalApp(t *testing.T) (*mocks.WithdrawalService, *fiber.App) {
	t.Helper()

	svc := mocks.NewWithdrawalService(t)
	h := handler.NewWithdrawalHandler(svc, handler.NewValidator())

	app := newApp()
	withdrawals := app.Group(withdrawalsPath, middleware.NewAuthMerchantMiddleware(testIssuer(t)))
	withdrawals.Post("/initiate", h.Initiate)
	withdrawals.Get("", h.List)
	withdrawals.Get("/:id", h.Get)

	return svc, app
}

func withdrawalFixture(merchantID uuid.UUID) *domain.Withdrawal {
	return &domain.Withdrawal{
		Transaction: domain.Transaction{
			ID:         uuid.New(),
			MerchantID: merchantID,
			Type:       domain.TransactionTypeWithdrawal,
			Amount:     "1000.00",
			Status:     domain.TransactionStatusPending,
			CreatedAt:  time.Now().UTC().Truncate(time.Second),
			UpdatedAt:  time.Now().UTC().Truncate(time.Second),
		},
		BankCode:      "SCB",
		AccountNumber: "1234567890",
		AccountName:   "Shop Owner",
	}
}

func validWithdrawalRequest() dto.InitiateWithdrawalRequest {
	return dto.InitiateWithdrawalRequest{
		Amount:        "1000.00",
		BankCode:      "SCB",
		AccountNumber: "1234567890",
		AccountName:   "Shop Owner",
	}
}

func TestWithdrawalHandlerInitiate(t *testing.T) {
	t.Run("returns 201 with a masked account number", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)
		merchantID := uuid.New()
		created := withdrawalFixture(merchantID)

		svc.EXPECT().
			Initiate(mock.Anything, mock.MatchedBy(func(actor domain.Actor) bool {
				return actor.Email == merchantEmail && actor.UserID == uuid.Nil
			}), domain.InitiateWithdrawal{
				MerchantID:    merchantID,
				Amount:        "1000.00",
				BankCode:      "SCB",
				AccountNumber: "1234567890",
				AccountName:   "Shop Owner",
			}).
			Return(created, nil).
			Once()

		res := doAuthorizedRequest(t, app, http.MethodPost, withdrawalsPath+"/initiate",
			merchantBearer(t, merchantID), validWithdrawalRequest())

		require.Equal(t, http.StatusCreated, res.StatusCode)

		body := decodeBody[dataEnvelope[dto.WithdrawalResponse]](t, res)
		assert.Equal(t, created.ID.String(), body.Data.TransactionID)
		assert.Equal(t, "SCB", body.Data.BankCode)
		assert.Equal(t, "xxxxxx7890", body.Data.AccountNumber)
		assert.NotContains(t, body.Data.AccountNumber, "123456")
	})

	t.Run("returns 401 without a bearer token", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)

		res := doRequest(t, app, http.MethodPost, withdrawalsPath+"/initiate", validWithdrawalRequest())

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		svc.AssertNotCalled(t, "Initiate")
	})

	t.Run("returns 401 for an admin token", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)

		res := doAuthorizedRequest(t, app, http.MethodPost, withdrawalsPath+"/initiate",
			bearerFor(t, uuid.New(), "admin@example.com", domain.TokenTypeAdmin), validWithdrawalRequest())

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		svc.AssertNotCalled(t, "Initiate")
	})

	t.Run("returns 400 for a missing required field", func(t *testing.T) {
		requests := []dto.InitiateWithdrawalRequest{
			{BankCode: "SCB", AccountNumber: "1234567890"},
			{Amount: "1000.00", AccountNumber: "1234567890"},
			{Amount: "1000.00", BankCode: "SCB"},
			{Amount: "abc", BankCode: "SCB", AccountNumber: "1234567890"},
		}

		for _, req := range requests {
			svc, app := setupWithdrawalApp(t)

			res := doAuthorizedRequest(t, app, http.MethodPost, withdrawalsPath+"/initiate",
				merchantBearer(t, uuid.New()), req)

			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			svc.AssertNotCalled(t, "Initiate")
		}
	})

	t.Run("maps domain rejections to their status codes", func(t *testing.T) {
		tests := map[error]int{
			domain.ErrInvalidAmount:        http.StatusBadRequest,
			domain.ErrInvalidBankCode:      http.StatusBadRequest,
			domain.ErrInvalidAccountNumber: http.StatusBadRequest,
			domain.ErrInsufficientBalance:  http.StatusBadRequest,
			domain.ErrWalletNotFound:       http.StatusNotFound,
			domain.ErrDuplicateReference:   http.StatusConflict,
			errDatabase:                    http.StatusInternalServerError,
		}

		for expectedErr, expectedStatus := range tests {
			svc, app := setupWithdrawalApp(t)

			svc.EXPECT().Initiate(mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedErr).Once()

			res := doAuthorizedRequest(t, app, http.MethodPost, withdrawalsPath+"/initiate",
				merchantBearer(t, uuid.New()), validWithdrawalRequest())

			assert.Equal(t, expectedStatus, res.StatusCode, expectedErr.Error())
		}
	})
}

func TestWithdrawalHandlerGet(t *testing.T) {
	t.Run("returns 200 with the withdrawal", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)
		merchantID := uuid.New()
		withdrawal := withdrawalFixture(merchantID)

		svc.EXPECT().Get(mock.Anything, merchantID, withdrawal.ID).Return(withdrawal, nil).Once()

		res := doAuthorizedRequest(t, app, http.MethodGet, withdrawalsPath+"/"+withdrawal.ID.String(),
			merchantBearer(t, merchantID), nil)

		require.Equal(t, http.StatusOK, res.StatusCode)

		body := decodeBody[dataEnvelope[dto.WithdrawalResponse]](t, res)
		assert.Equal(t, withdrawal.ID.String(), body.Data.TransactionID)
		assert.Equal(t, "xxxxxx7890", body.Data.AccountNumber)
	})

	t.Run("returns 400 for a non-uuid id", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)

		res := doAuthorizedRequest(t, app, http.MethodGet, withdrawalsPath+"/"+invalidUUID,
			merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		svc.AssertNotCalled(t, "Get")
	})

	t.Run("returns 404 for another merchant's withdrawal", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)
		merchantID := uuid.New()
		id := uuid.New()

		svc.EXPECT().Get(mock.Anything, merchantID, id).Return(nil, domain.ErrTransactionNotFound).Once()

		res := doAuthorizedRequest(t, app, http.MethodGet, withdrawalsPath+"/"+id.String(),
			merchantBearer(t, merchantID), nil)

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestWithdrawalHandlerList(t *testing.T) {
	t.Run("returns 200 with pagination scoped to the merchant", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)
		merchantID := uuid.New()
		withdrawals := []domain.Withdrawal{*withdrawalFixture(merchantID)}

		svc.EXPECT().
			List(mock.Anything, domain.WithdrawalQuery{
				ListQuery:  domain.ListQuery{Page: 2, Size: 5, SortBy: "amount", OrderBy: "asc"},
				MerchantID: merchantID,
			}).
			Return(withdrawals, int64(6), nil).
			Once()

		res := doAuthorizedRequest(t, app, http.MethodGet,
			withdrawalsPath+"?page=2&size=5&sort_by=amount&order_by=asc", merchantBearer(t, merchantID), nil)

		require.Equal(t, http.StatusOK, res.StatusCode)

		body := decodeBody[paginateEnvelope[[]dto.WithdrawalListItemResponse]](t, res)
		require.Len(t, body.Data, 1)
		assert.Equal(t, "xxxxxx7890", body.Data[0].AccountNumber)
		assert.Equal(t, int64(6), body.Pagination.TotalItem)
	})

	t.Run("returns 400 for an out-of-range page size", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)

		res := doAuthorizedRequest(t, app, http.MethodGet, withdrawalsPath+"?size=101",
			merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		svc.AssertNotCalled(t, "List")
	})

	t.Run("returns 400 for an unknown order direction", func(t *testing.T) {
		svc, app := setupWithdrawalApp(t)

		res := doAuthorizedRequest(t, app, http.MethodGet, withdrawalsPath+"?order_by=sideways",
			merchantBearer(t, uuid.New()), nil)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		svc.AssertNotCalled(t, "List")
	})
}
