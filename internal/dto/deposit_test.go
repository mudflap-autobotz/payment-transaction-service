package dto_test

import (
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deposit(status string, qrExpiryAt time.Time) domain.Deposit {
	return domain.Deposit{
		Transaction: domain.Transaction{
			ID:                uuid.New(),
			MerchantID:        uuid.New(),
			Type:              domain.TransactionTypeDeposit,
			Amount:            "500.00",
			Status:            status,
			MerchantReference: "ORDER-1",
			BankReference:     "BANK-1",
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		},
		QRCode:     "00020101021229370016A000000677010111",
		QRType:     domain.QRTypePromptPay,
		QRExpiryAt: qrExpiryAt,
	}
}

func TestToDepositResponse(t *testing.T) {
	t.Run("maps every field", func(t *testing.T) {
		source := deposit(domain.TransactionStatusPending, time.Now().UTC().Add(time.Hour))
		confirmedAt := time.Now().UTC()
		source.PaymentConfirmedAt = &confirmedAt

		result := dto.ToDepositResponse(&source)

		assert.Equal(t, source.ID.String(), result.TransactionID)
		assert.Equal(t, source.MerchantID.String(), result.MerchantID)
		assert.Equal(t, source.Amount, result.Amount)
		assert.Equal(t, source.Status, result.Status)
		assert.Equal(t, source.QRCode, result.QRCode)
		assert.Equal(t, source.QRType, result.QRType)
		assert.Equal(t, source.QRExpiryAt, result.ExpiresAt)
		assert.Equal(t, source.MerchantReference, result.MerchantReference)
		assert.Equal(t, source.BankReference, result.BankReference)
		require.NotNil(t, result.PaymentConfirmedAt)
		assert.Equal(t, confirmedAt, *result.PaymentConfirmedAt)
	})

	t.Run("flags a pending deposit past its qr expiry as expired", func(t *testing.T) {
		source := deposit(domain.TransactionStatusPending, time.Now().UTC().Add(-time.Minute))

		assert.True(t, dto.ToDepositResponse(&source).Expired)
	})

	t.Run("does not flag a pending deposit still inside its window", func(t *testing.T) {
		source := deposit(domain.TransactionStatusPending, time.Now().UTC().Add(time.Hour))

		assert.False(t, dto.ToDepositResponse(&source).Expired)
	})

	t.Run("does not flag a settled deposit past its qr expiry", func(t *testing.T) {
		source := deposit(domain.TransactionStatusCompleted, time.Now().UTC().Add(-time.Hour))

		assert.False(t, dto.ToDepositResponse(&source).Expired)
	})
}

func TestToDepositListItemResponses(t *testing.T) {
	t.Run("maps each deposit and keeps the order", func(t *testing.T) {
		deposits := []domain.Deposit{
			deposit(domain.TransactionStatusPending, time.Now().UTC().Add(time.Hour)),
			deposit(domain.TransactionStatusCompleted, time.Now().UTC().Add(-time.Hour)),
		}

		result := dto.ToDepositListItemResponses(deposits)

		require.Len(t, result, 2)
		assert.Equal(t, deposits[0].ID.String(), result[0].TransactionID)
		assert.Equal(t, deposits[1].ID.String(), result[1].TransactionID)
		assert.False(t, result[0].Expired)
		assert.False(t, result[1].Expired)
	})

	t.Run("returns an empty slice rather than nil", func(t *testing.T) {
		result := dto.ToDepositListItemResponses(nil)

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}
