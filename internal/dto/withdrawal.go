package dto

import (
	"strings"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
)

const maskedAccountSuffixLength = 4

type InitiateWithdrawalRequest struct {
	Amount            string `json:"amount" validate:"required,numeric"`
	BankCode          string `json:"bank_code" validate:"required,max=20"`
	AccountNumber     string `json:"account_number" validate:"required,max=50"`
	AccountName       string `json:"account_name" validate:"omitempty,max=255"`
	MerchantReference string `json:"merchant_reference" validate:"omitempty,max=100"`
}

type WithdrawalQuery struct {
	PaginationQuery
	DateRangeQuery

	Status string `query:"status" validate:"omitempty,oneof=PENDING SUBMITTED UNCONFIRMED COMPLETED FAILED EXPIRED CANCELLED"`
}

type WithdrawalResponse struct {
	TransactionID     string     `json:"transaction_id"`
	MerchantID        string     `json:"merchant_id"`
	Amount            string     `json:"amount"`
	Status            string     `json:"status"`
	BankCode          string     `json:"bank_code"`
	AccountNumber     string     `json:"account_number"`
	AccountName       string     `json:"account_name,omitempty"`
	MerchantReference string     `json:"merchant_reference,omitempty"`
	BankReference     string     `json:"bank_reference,omitempty"`
	SubmittedAt       *time.Time `json:"submitted_at"`
	ConfirmedAt       *time.Time `json:"confirmed_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type WithdrawalListItemResponse struct {
	TransactionID     string     `json:"transaction_id"`
	Amount            string     `json:"amount"`
	Status            string     `json:"status"`
	BankCode          string     `json:"bank_code"`
	AccountNumber     string     `json:"account_number"`
	MerchantReference string     `json:"merchant_reference,omitempty"`
	SubmittedAt       *time.Time `json:"submitted_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

func ToWithdrawalResponse(withdrawal *domain.Withdrawal) WithdrawalResponse {
	return WithdrawalResponse{
		TransactionID:     withdrawal.ID.String(),
		MerchantID:        withdrawal.MerchantID.String(),
		Amount:            withdrawal.Amount,
		Status:            withdrawal.Status,
		BankCode:          withdrawal.BankCode,
		AccountNumber:     MaskAccountNumber(withdrawal.AccountNumber),
		AccountName:       withdrawal.AccountName,
		MerchantReference: withdrawal.MerchantReference,
		BankReference:     withdrawal.BankReference,
		SubmittedAt:       withdrawal.SubmittedAt,
		ConfirmedAt:       withdrawal.ConfirmedAt,
		CreatedAt:         withdrawal.CreatedAt,
		UpdatedAt:         withdrawal.UpdatedAt,
	}
}

func ToWithdrawalListItemResponses(withdrawals []domain.Withdrawal) []WithdrawalListItemResponse {
	responses := make([]WithdrawalListItemResponse, 0, len(withdrawals))

	for i := range withdrawals {
		responses = append(responses, WithdrawalListItemResponse{
			TransactionID:     withdrawals[i].ID.String(),
			Amount:            withdrawals[i].Amount,
			Status:            withdrawals[i].Status,
			BankCode:          withdrawals[i].BankCode,
			AccountNumber:     MaskAccountNumber(withdrawals[i].AccountNumber),
			MerchantReference: withdrawals[i].MerchantReference,
			SubmittedAt:       withdrawals[i].SubmittedAt,
			CreatedAt:         withdrawals[i].CreatedAt,
		})
	}

	return responses
}

func MaskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= maskedAccountSuffixLength {
		return accountNumber
	}

	visible := accountNumber[len(accountNumber)-maskedAccountSuffixLength:]

	return strings.Repeat("x", len(accountNumber)-maskedAccountSuffixLength) + visible
}
