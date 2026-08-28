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

func withdrawal() domain.Withdrawal {
	return domain.Withdrawal{
		Transaction: domain.Transaction{
			ID:                uuid.New(),
			MerchantID:        uuid.New(),
			Type:              domain.TransactionTypeWithdrawal,
			Amount:            "1000.00",
			Status:            domain.TransactionStatusPending,
			MerchantReference: "PAYOUT-1",
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		},
		BankCode:      "SCB",
		AccountNumber: "1234567890",
		AccountName:   "Shop Owner",
	}
}

func TestToWithdrawalResponse(t *testing.T) {
	t.Run("maps every field and masks the account number", func(t *testing.T) {
		source := withdrawal()
		submittedAt := time.Now().UTC()
		source.SubmittedAt = &submittedAt

		result := dto.ToWithdrawalResponse(&source)

		assert.Equal(t, source.ID.String(), result.TransactionID)
		assert.Equal(t, source.MerchantID.String(), result.MerchantID)
		assert.Equal(t, source.Amount, result.Amount)
		assert.Equal(t, source.BankCode, result.BankCode)
		assert.Equal(t, "xxxxxx7890", result.AccountNumber)
		assert.Equal(t, source.AccountName, result.AccountName)
		assert.Equal(t, source.MerchantReference, result.MerchantReference)
		require.NotNil(t, result.SubmittedAt)
		assert.Equal(t, submittedAt, *result.SubmittedAt)
	})
}

func TestToWithdrawalListItemResponses(t *testing.T) {
	t.Run("masks every account number", func(t *testing.T) {
		withdrawals := []domain.Withdrawal{withdrawal(), withdrawal()}

		result := dto.ToWithdrawalListItemResponses(withdrawals)

		require.Len(t, result, 2)
		for i := range result {
			assert.Equal(t, "xxxxxx7890", result[i].AccountNumber)
			assert.Equal(t, withdrawals[i].ID.String(), result[i].TransactionID)
		}
	})

	t.Run("returns an empty slice rather than nil", func(t *testing.T) {
		result := dto.ToWithdrawalListItemResponses(nil)

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func TestMaskAccountNumber(t *testing.T) {
	tests := map[string]string{
		"1234567890":      "xxxxxx7890",
		"123456789012345": "xxxxxxxxxxx2345",
		"7890":            "7890",
		"890":             "890",
		"":                "",
	}

	for accountNumber, expected := range tests {
		assert.Equal(t, expected, dto.MaskAccountNumber(accountNumber), accountNumber)
	}
}
