package transaction

import (
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type TransactionDB struct {
	bun.BaseModel `bun:"table:transactions,alias:t"`

	ID                uuid.UUID  `bun:"id,pk,nullzero,default:uuidv7()"`
	MerchantID        uuid.UUID  `bun:"merchant_id"`
	Type              string     `bun:"type"`
	Amount            string     `bun:"amount"`
	Status            string     `bun:"status,nullzero"`
	MerchantReference string     `bun:"merchant_reference,nullzero"`
	BankReference     string     `bun:"bank_reference,nullzero"`
	Settled           bool       `bun:"settled"`
	SettledAt         *time.Time `bun:"settled_at"`
	CreatedAt         time.Time  `bun:"created_at,nullzero,default:now()"`
	UpdatedAt         time.Time  `bun:"updated_at,nullzero,default:now()"`
}

type DepositDB struct {
	bun.BaseModel `bun:"table:deposits,alias:d"`

	ID                 uuid.UUID  `bun:"id,pk,nullzero,default:uuidv7()"`
	TransactionID      uuid.UUID  `bun:"transaction_id"`
	QRCode             string     `bun:"qr_code"`
	QRType             string     `bun:"qr_type,nullzero"`
	QRExpiryAt         time.Time  `bun:"qr_expiry_at"`
	PaymentConfirmedAt *time.Time `bun:"payment_confirmed_at"`
}

type WithdrawalDB struct {
	bun.BaseModel `bun:"table:withdrawals,alias:wd"`

	ID            uuid.UUID  `bun:"id,pk,nullzero,default:uuidv7()"`
	TransactionID uuid.UUID  `bun:"transaction_id"`
	BankCode      string     `bun:"bank_code"`
	AccountNumber string     `bun:"account_number"`
	AccountName   string     `bun:"account_name,nullzero"`
	SubmittedAt   *time.Time `bun:"submitted_at"`
	ConfirmedAt   *time.Time `bun:"confirmed_at"`
}

type WalletDB struct {
	bun.BaseModel `bun:"table:wallets,alias:w"`

	ID         uuid.UUID `bun:"id,pk,nullzero,default:uuidv7()"`
	MerchantID uuid.UUID `bun:"merchant_id"`
	Balance    string    `bun:"balance,nullzero"`
	Reserved   string    `bun:"reserved,nullzero"`
	CreatedAt  time.Time `bun:"created_at,nullzero,default:now()"`
	UpdatedAt  time.Time `bun:"updated_at,nullzero,default:now()"`
}

type depositRow struct {
	TransactionDB

	QRCode             string     `bun:"qr_code"`
	QRType             string     `bun:"qr_type"`
	QRExpiryAt         time.Time  `bun:"qr_expiry_at"`
	PaymentConfirmedAt *time.Time `bun:"payment_confirmed_at"`
}

type withdrawalRow struct {
	TransactionDB

	BankCode      string     `bun:"bank_code"`
	AccountNumber string     `bun:"account_number"`
	AccountName   string     `bun:"account_name"`
	SubmittedAt   *time.Time `bun:"submitted_at"`
	ConfirmedAt   *time.Time `bun:"confirmed_at"`
}

func (m TransactionDB) toDomain() domain.Transaction {
	return domain.Transaction{
		ID:                m.ID,
		MerchantID:        m.MerchantID,
		Type:              m.Type,
		Amount:            m.Amount,
		Status:            m.Status,
		MerchantReference: m.MerchantReference,
		BankReference:     m.BankReference,
		Settled:           m.Settled,
		SettledAt:         m.SettledAt,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func (r depositRow) ToDomain() domain.Deposit {
	return domain.Deposit{
		Transaction:        r.TransactionDB.toDomain(),
		QRCode:             r.QRCode,
		QRType:             r.QRType,
		QRExpiryAt:         r.QRExpiryAt,
		PaymentConfirmedAt: r.PaymentConfirmedAt,
	}
}

func (r withdrawalRow) ToDomain() domain.Withdrawal {
	return domain.Withdrawal{
		Transaction:   r.TransactionDB.toDomain(),
		BankCode:      r.BankCode,
		AccountNumber: r.AccountNumber,
		AccountName:   r.AccountName,
		SubmittedAt:   r.SubmittedAt,
		ConfirmedAt:   r.ConfirmedAt,
	}
}

func toDepositDomain(transactionDB *TransactionDB, depositDB *DepositDB) domain.Deposit {
	return domain.Deposit{
		Transaction:        transactionDB.toDomain(),
		QRCode:             depositDB.QRCode,
		QRType:             depositDB.QRType,
		QRExpiryAt:         depositDB.QRExpiryAt,
		PaymentConfirmedAt: depositDB.PaymentConfirmedAt,
	}
}

func toWithdrawalDomain(transactionDB *TransactionDB, withdrawalDB *WithdrawalDB) domain.Withdrawal {
	return domain.Withdrawal{
		Transaction:   transactionDB.toDomain(),
		BankCode:      withdrawalDB.BankCode,
		AccountNumber: withdrawalDB.AccountNumber,
		AccountName:   withdrawalDB.AccountName,
		SubmittedAt:   withdrawalDB.SubmittedAt,
		ConfirmedAt:   withdrawalDB.ConfirmedAt,
	}
}
