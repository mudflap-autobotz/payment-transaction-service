package transaction

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mudflap-autobotz/payment-common/database"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const zeroAmount = "0.00"

type summaryTotalsRow struct {
	TodayDepositTotal    string `bun:"today_deposit_total"`
	TodayDepositCount    int64  `bun:"today_deposit_count"`
	TodayWithdrawalTotal string `bun:"today_withdrawal_total"`
	TodayWithdrawalCount int64  `bun:"today_withdrawal_count"`
	PendingDepositCount  int64  `bun:"pending_deposit_count"`
}

type SummaryRepository struct {
	readDB *database.ReadDB
	tracer trace.Tracer
}

func NewSummaryRepository(readDB *database.ReadDB) *SummaryRepository {
	return &SummaryRepository{
		readDB: readDB,
		tracer: otel.Tracer("repository.summary"),
	}
}

func (r SummaryRepository) MerchantSummary(ctx context.Context, merchantID uuid.UUID, since time.Time) (*domain.MerchantSummary, error) {
	ctx, span := r.tracer.Start(ctx, "SummaryRepository.MerchantSummary")
	defer span.End()

	walletDB := new(WalletDB)

	err := r.readDB.NewSelect().
		Model(walletDB).
		Where("w.merchant_id = ?", merchantID).
		Scan(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	totals := new(summaryTotalsRow)

	err = r.readDB.NewSelect().
		Model((*TransactionDB)(nil)).
		ColumnExpr("COALESCE(SUM(amount) FILTER (WHERE type = ? AND status = ? AND created_at >= ?), 0)::numeric(15,2)::text AS today_deposit_total",
			domain.TransactionTypeDeposit, domain.TransactionStatusCompleted, since).
		ColumnExpr("COUNT(*) FILTER (WHERE type = ? AND status = ? AND created_at >= ?) AS today_deposit_count",
			domain.TransactionTypeDeposit, domain.TransactionStatusCompleted, since).
		ColumnExpr("COALESCE(SUM(amount) FILTER (WHERE type = ? AND status = ? AND created_at >= ?), 0)::numeric(15,2)::text AS today_withdrawal_total",
			domain.TransactionTypeWithdrawal, domain.TransactionStatusCompleted, since).
		ColumnExpr("COUNT(*) FILTER (WHERE type = ? AND status = ? AND created_at >= ?) AS today_withdrawal_count",
			domain.TransactionTypeWithdrawal, domain.TransactionStatusCompleted, since).
		ColumnExpr("COUNT(*) FILTER (WHERE type = ? AND status = ?) AS pending_deposit_count",
			domain.TransactionTypeDeposit, domain.TransactionStatusPending).
		Where("merchant_id = ?", merchantID).
		Scan(ctx, totals)
	if err != nil {
		return nil, err
	}

	return &domain.MerchantSummary{
		Balance:              defaultAmount(walletDB.Balance),
		Reserved:             defaultAmount(walletDB.Reserved),
		TodayDepositTotal:    defaultAmount(totals.TodayDepositTotal),
		TodayDepositCount:    totals.TodayDepositCount,
		TodayWithdrawalTotal: defaultAmount(totals.TodayWithdrawalTotal),
		TodayWithdrawalCount: totals.TodayWithdrawalCount,
		PendingDepositCount:  totals.PendingDepositCount,
	}, nil
}

func defaultAmount(value string) string {
	if value == "" {
		return zeroAmount
	}

	return value
}
