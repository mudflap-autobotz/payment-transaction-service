package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MerchantSummary struct {
	Balance              string
	Reserved             string
	TodayDepositTotal    string
	TodayDepositCount    int64
	TodayWithdrawalTotal string
	TodayWithdrawalCount int64
	PendingDepositCount  int64
}

type SummaryRepository interface {
	MerchantSummary(ctx context.Context, merchantID uuid.UUID, since time.Time) (*MerchantSummary, error)
}

type SummaryService interface {
	Merchant(ctx context.Context, merchantID uuid.UUID) (*MerchantSummary, error)
}
