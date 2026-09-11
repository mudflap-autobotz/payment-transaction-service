package dto

import (
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
)

type InitiateDepositRequest struct {
	Amount            string `json:"amount" validate:"required,numeric"`
	MerchantReference string `json:"merchant_reference" validate:"omitempty,max=100"`
}

type DepositQuery struct {
	PaginationQuery
	DateRangeQuery

	Status string `query:"status" validate:"omitempty,oneof=PENDING SUBMITTED UNCONFIRMED COMPLETED FAILED EXPIRED CANCELLED"`
}

type DepositResponse struct {
	TransactionID      string     `json:"transaction_id"`
	MerchantID         string     `json:"merchant_id"`
	Amount             string     `json:"amount"`
	Status             string     `json:"status"`
	QRCode             string     `json:"qr_code"`
	QRType             string     `json:"qr_type"`
	ExpiresAt          time.Time  `json:"expires_at"`
	Expired            bool       `json:"expired"`
	MerchantReference  string     `json:"merchant_reference,omitempty"`
	BankReference      string     `json:"bank_reference,omitempty"`
	PaymentConfirmedAt *time.Time `json:"payment_confirmed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type DepositListItemResponse struct {
	TransactionID      string     `json:"transaction_id"`
	Amount             string     `json:"amount"`
	Status             string     `json:"status"`
	ExpiresAt          time.Time  `json:"expires_at"`
	Expired            bool       `json:"expired"`
	MerchantReference  string     `json:"merchant_reference,omitempty"`
	PaymentConfirmedAt *time.Time `json:"payment_confirmed_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

func ToDepositResponse(deposit *domain.Deposit) DepositResponse {
	return DepositResponse{
		TransactionID:      deposit.ID.String(),
		MerchantID:         deposit.MerchantID.String(),
		Amount:             deposit.Amount,
		Status:             deposit.Status,
		QRCode:             deposit.QRCode,
		QRType:             deposit.QRType,
		ExpiresAt:          deposit.QRExpiryAt,
		Expired:            isDepositExpired(deposit),
		MerchantReference:  deposit.MerchantReference,
		BankReference:      deposit.BankReference,
		PaymentConfirmedAt: deposit.PaymentConfirmedAt,
		CreatedAt:          deposit.CreatedAt,
		UpdatedAt:          deposit.UpdatedAt,
	}
}

func ToDepositListItemResponses(deposits []domain.Deposit) []DepositListItemResponse {
	responses := make([]DepositListItemResponse, 0, len(deposits))

	for i := range deposits {
		responses = append(responses, DepositListItemResponse{
			TransactionID:      deposits[i].ID.String(),
			Amount:             deposits[i].Amount,
			Status:             deposits[i].Status,
			ExpiresAt:          deposits[i].QRExpiryAt,
			Expired:            isDepositExpired(&deposits[i]),
			MerchantReference:  deposits[i].MerchantReference,
			PaymentConfirmedAt: deposits[i].PaymentConfirmedAt,
			CreatedAt:          deposits[i].CreatedAt,
		})
	}

	return responses
}

func isDepositExpired(deposit *domain.Deposit) bool {
	return deposit.Status == domain.TransactionStatusPending && time.Now().After(deposit.QRExpiryAt)
}
