package kafka

import "time"

type DepositCreatedEvent struct {
	TransactionID     string    `json:"transaction_id"`
	MerchantID        string    `json:"merchant_id"`
	Amount            string    `json:"amount"`
	Status            string    `json:"status"`
	QRCode            string    `json:"qr_code"`
	QRExpiryAt        time.Time `json:"qr_expiry_at"`
	MerchantReference string    `json:"merchant_reference,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type WithdrawalCreatedEvent struct {
	TransactionID     string    `json:"transaction_id"`
	MerchantID        string    `json:"merchant_id"`
	Amount            string    `json:"amount"`
	Status            string    `json:"status"`
	BankCode          string    `json:"bank_code"`
	AccountNumber     string    `json:"account_number"`
	AccountName       string    `json:"account_name,omitempty"`
	MerchantReference string    `json:"merchant_reference,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type envelope struct {
	Event      string `json:"event"`
	OccurredAt string `json:"occurred_at"`
	Data       any    `json:"data"`
}
