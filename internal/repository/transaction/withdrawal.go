package transaction

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mudflap-autobotz/payment-common/database"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/repository/audit"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/repository/pgerr"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const withdrawalColumns = "wd.bank_code, wd.account_number, wd.account_name, wd.submitted_at, wd.confirmed_at"

type WithdrawalRepository struct {
	writeDB *database.WriteDB
	readDB  *database.ReadDB
	tracer  trace.Tracer
}

func NewWithdrawalRepository(writeDB *database.WriteDB, readDB *database.ReadDB) *WithdrawalRepository {
	return &WithdrawalRepository{
		writeDB: writeDB,
		readDB:  readDB,
		tracer:  otel.Tracer("repository.withdrawal"),
	}
}

func (r WithdrawalRepository) Create(ctx context.Context, input domain.CreateWithdrawal, auditEntry *domain.AuditEntry) (*domain.Withdrawal, error) {
	ctx, span := r.tracer.Start(ctx, "WithdrawalRepository.Create")
	defer span.End()

	var created domain.Withdrawal

	err := r.writeDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := reserveWalletBalance(ctx, tx, input.MerchantID, input.Amount); err != nil {
			return err
		}

		transactionDB := &TransactionDB{
			MerchantID:        input.MerchantID,
			Type:              domain.TransactionTypeWithdrawal,
			Amount:            input.Amount,
			Status:            domain.TransactionStatusPending,
			MerchantReference: input.MerchantReference,
		}

		if _, err := tx.NewInsert().Model(transactionDB).Returning("*").Exec(ctx); err != nil {
			if pgerr.IsUniqueViolation(err) {
				return domain.ErrDuplicateReference
			}

			return err
		}

		withdrawalDB := &WithdrawalDB{
			TransactionID: transactionDB.ID,
			BankCode:      input.BankCode,
			AccountNumber: input.AccountNumber,
			AccountName:   input.AccountName,
		}

		if _, err := tx.NewInsert().Model(withdrawalDB).
			Returning("id, submitted_at, confirmed_at").
			Exec(ctx); err != nil {
			return err
		}

		if auditEntry != nil {
			auditEntry.EntityID = transactionDB.ID
		}

		if err := audit.Insert(ctx, tx, auditEntry); err != nil {
			return err
		}

		created = toWithdrawalDomain(transactionDB, withdrawalDB)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func reserveWalletBalance(ctx context.Context, tx bun.Tx, merchantID uuid.UUID, amount string) error {
	walletDB := new(WalletDB)

	err := tx.NewSelect().
		Model(walletDB).
		Where("w.merchant_id = ?", merchantID).
		For("UPDATE").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrWalletNotFound
		}

		return err
	}

	result, err := tx.NewUpdate().
		Model((*WalletDB)(nil)).
		Set("reserved = w.reserved + ?::numeric", amount).
		Set("updated_at = now()").
		Where("w.merchant_id = ?", merchantID).
		Where("w.reserved + ?::numeric <= w.balance", amount).
		Exec(ctx)
	if err != nil {
		if pgerr.IsCheckViolation(err) {
			return domain.ErrInsufficientBalance
		}

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrInsufficientBalance
	}

	return nil
}

func (r WithdrawalRepository) GetByID(ctx context.Context, merchantID, id uuid.UUID) (*domain.Withdrawal, error) {
	ctx, span := r.tracer.Start(ctx, "WithdrawalRepository.GetByID")
	defer span.End()

	row := new(withdrawalRow)

	err := r.readDB.NewSelect().
		Model((*TransactionDB)(nil)).
		ColumnExpr("t.*").
		ColumnExpr(withdrawalColumns).
		Join("JOIN withdrawals AS wd ON wd.transaction_id = t.id").
		Where("t.id = ?", id).
		Where("t.merchant_id = ?", merchantID).
		Where("t.type = ?", domain.TransactionTypeWithdrawal).
		Scan(ctx, row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}

		return nil, err
	}

	withdrawal := row.ToDomain()

	return &withdrawal, nil
}

func (r WithdrawalRepository) List(ctx context.Context, query domain.WithdrawalQuery) ([]domain.Withdrawal, int64, error) {
	ctx, span := r.tracer.Start(ctx, "WithdrawalRepository.List")
	defer span.End()

	rows := make([]withdrawalRow, 0)

	selectQuery := r.readDB.NewSelect().
		Model((*TransactionDB)(nil)).
		ColumnExpr("t.*").
		ColumnExpr(withdrawalColumns).
		Join("JOIN withdrawals AS wd ON wd.transaction_id = t.id").
		Where("t.merchant_id = ?", query.MerchantID).
		Where("t.type = ?", domain.TransactionTypeWithdrawal)

	if query.Status != "" {
		selectQuery = selectQuery.Where("t.status = ?", query.Status)
	}

	totalItem, err := selectQuery.
		Order("t."+query.SortBy+" "+query.OrderBy).
		Limit(query.Size).
		Offset(query.Offset()).
		ScanAndCount(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	withdrawals := make([]domain.Withdrawal, 0, len(rows))
	for i := range rows {
		withdrawals = append(withdrawals, rows[i].ToDomain())
	}

	return withdrawals, int64(totalItem), nil
}
