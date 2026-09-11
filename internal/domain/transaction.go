package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	TransactionTypeDeposit    = "DEPOSIT"
	TransactionTypeWithdrawal = "WITHDRAWAL"

	TransactionStatusPending     = "PENDING"
	TransactionStatusSubmitted   = "SUBMITTED"
	TransactionStatusUnconfirmed = "UNCONFIRMED"
	TransactionStatusCompleted   = "COMPLETED"
	TransactionStatusFailed      = "FAILED"
	TransactionStatusExpired     = "EXPIRED"
	TransactionStatusCancelled   = "CANCELLED"

	MaxMerchantReferenceLength = 100
)

var (
	ErrTransactionNotFound  = NotFoundError("transaction")
	ErrWalletNotFound       = NotFoundError("wallet")
	ErrInsufficientBalance  = InvalidInputError("insufficient balance")
	ErrInvalidAmount        = InvalidInputError("amount is out of the allowed range")
	ErrInvalidBankCode      = InvalidInputError("bank code is not supported")
	ErrInvalidAccountNumber = InvalidInputError("account number format is invalid")
	ErrDuplicateReference   = ConflictReasonError("merchant_reference already used")
)

type Transaction struct {
	ID                uuid.UUID
	MerchantID        uuid.UUID
	Type              string
	Amount            string
	Status            string
	MerchantReference string
	BankReference     string
	Settled           bool
	SettledAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
