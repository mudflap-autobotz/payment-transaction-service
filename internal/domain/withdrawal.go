package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Withdrawal struct {
	Transaction

	BankCode      string
	AccountNumber string
	AccountName   string
	SubmittedAt   *time.Time
	ConfirmedAt   *time.Time
}

type WithdrawalQuery struct {
	ListQuery

	MerchantID uuid.UUID
	Status     string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type CreateWithdrawal struct {
	MerchantID        uuid.UUID
	Amount            string
	MerchantReference string
	BankCode          string
	AccountNumber     string
	AccountName       string
}

type InitiateWithdrawal struct {
	MerchantID        uuid.UUID
	Amount            string
	MerchantReference string
	BankCode          string
	AccountNumber     string
	AccountName       string
}

type WithdrawalRepository interface {
	Create(ctx context.Context, input CreateWithdrawal, audit *AuditEntry) (*Withdrawal, error)
	GetByID(ctx context.Context, merchantID, id uuid.UUID) (*Withdrawal, error)
	List(ctx context.Context, query WithdrawalQuery) ([]Withdrawal, int64, error)
}

type WithdrawalService interface {
	Initiate(ctx context.Context, actor Actor, input InitiateWithdrawal) (*Withdrawal, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*Withdrawal, error)
	List(ctx context.Context, query WithdrawalQuery) ([]Withdrawal, int64, error)
}
