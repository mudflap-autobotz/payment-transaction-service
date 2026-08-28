package domain

import "github.com/google/uuid"

type Wallet struct {
	ID         uuid.UUID
	MerchantID uuid.UUID
	Balance    string
	Reserved   string
}
