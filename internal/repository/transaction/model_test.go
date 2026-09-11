package transaction

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalRowToDomainCarriesTheSourceAccountAndBankResponse(t *testing.T) {
	sourceBankAccountID := uuid.New()

	row := withdrawalRow{
		BankCode:            "KBANK",
		AccountNumber:       "1234567890",
		AccountName:         "Shop Owner",
		SourceBankAccountID: &sourceBankAccountID,
		BankResponse:        json.RawMessage(`{"outcome":"ACCEPTED","provider_ref":"REF1"}`),
	}

	result := row.ToDomain()

	require.NotNil(t, result.SourceBankAccountID)
	assert.Equal(t, sourceBankAccountID, *result.SourceBankAccountID)
	assert.JSONEq(t, `{"outcome":"ACCEPTED","provider_ref":"REF1"}`, string(result.BankResponse))
	assert.Equal(t, "KBANK", result.BankCode)
}

func TestWithdrawalRowToDomainLeavesAnUnsettledWithdrawalEmpty(t *testing.T) {
	result := withdrawalRow{BankCode: "SCB"}.ToDomain()

	assert.Nil(t, result.SourceBankAccountID)
	assert.Nil(t, result.BankResponse)
}

func TestToWithdrawalDomainCarriesTheSourceAccountAndBankResponse(t *testing.T) {
	sourceBankAccountID := uuid.New()

	withdrawalDB := &WithdrawalDB{
		BankCode:            "KBANK",
		SourceBankAccountID: &sourceBankAccountID,
		BankResponse:        json.RawMessage(`{"outcome":"UNKNOWN"}`),
	}

	result := toWithdrawalDomain(&TransactionDB{ID: uuid.New()}, withdrawalDB)

	require.NotNil(t, result.SourceBankAccountID)
	assert.Equal(t, sourceBankAccountID, *result.SourceBankAccountID)
	assert.JSONEq(t, `{"outcome":"UNKNOWN"}`, string(result.BankResponse))
}
