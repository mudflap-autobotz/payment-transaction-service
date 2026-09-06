package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const QRTypePromptPay = "PROMPTPAY"

type Deposit struct {
	Transaction

	QRCode             string
	QRType             string
	QRExpiryAt         time.Time
	PaymentConfirmedAt *time.Time
}

type DepositQuery struct {
	ListQuery

	MerchantID uuid.UUID
	Status     string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type CreateDeposit struct {
	MerchantID        uuid.UUID
	Amount            string
	MerchantReference string
	QRCode            string
	QRType            string
	QRExpiryAt        time.Time
}

type InitiateDeposit struct {
	MerchantID        uuid.UUID
	Amount            string
	MerchantReference string
}

type DepositRepository interface {
	Create(ctx context.Context, input CreateDeposit, audit *AuditEntry) (*Deposit, error)
	GetByID(ctx context.Context, merchantID, id uuid.UUID) (*Deposit, error)
	List(ctx context.Context, query DepositQuery) ([]Deposit, int64, error)
}

type DepositService interface {
	Initiate(ctx context.Context, actor Actor, input InitiateDeposit) (*Deposit, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*Deposit, error)
	List(ctx context.Context, query DepositQuery) ([]Deposit, int64, error)
}
