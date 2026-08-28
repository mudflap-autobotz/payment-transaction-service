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

const depositColumns = "d.qr_code, d.qr_type, d.qr_expiry_at, d.payment_confirmed_at"

type DepositRepository struct {
	writeDB *database.WriteDB
	readDB  *database.ReadDB
	tracer  trace.Tracer
}

func NewDepositRepository(writeDB *database.WriteDB, readDB *database.ReadDB) *DepositRepository {
	return &DepositRepository{
		writeDB: writeDB,
		readDB:  readDB,
		tracer:  otel.Tracer("repository.deposit"),
	}
}

func (r DepositRepository) Create(ctx context.Context, input domain.CreateDeposit, auditEntry *domain.AuditEntry) (*domain.Deposit, error) {
	ctx, span := r.tracer.Start(ctx, "DepositRepository.Create")
	defer span.End()

	var created domain.Deposit

	err := r.writeDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		transactionDB := &TransactionDB{
			MerchantID:        input.MerchantID,
			Type:              domain.TransactionTypeDeposit,
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

		depositDB := &DepositDB{
			TransactionID: transactionDB.ID,
			QRCode:        input.QRCode,
			QRType:        input.QRType,
			QRExpiryAt:    input.QRExpiryAt,
		}

		if _, err := tx.NewInsert().Model(depositDB).
			Returning("id, qr_type, payment_confirmed_at").
			Exec(ctx); err != nil {
			return err
		}

		if auditEntry != nil {
			auditEntry.EntityID = transactionDB.ID
		}

		if err := audit.Insert(ctx, tx, auditEntry); err != nil {
			return err
		}

		created = toDepositDomain(transactionDB, depositDB)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r DepositRepository) GetByID(ctx context.Context, merchantID, id uuid.UUID) (*domain.Deposit, error) {
	ctx, span := r.tracer.Start(ctx, "DepositRepository.GetByID")
	defer span.End()

	row := new(depositRow)

	err := r.readDB.NewSelect().
		Model((*TransactionDB)(nil)).
		ColumnExpr("t.*").
		ColumnExpr(depositColumns).
		Join("JOIN deposits AS d ON d.transaction_id = t.id").
		Where("t.id = ?", id).
		Where("t.merchant_id = ?", merchantID).
		Where("t.type = ?", domain.TransactionTypeDeposit).
		Scan(ctx, row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}

		return nil, err
	}

	deposit := row.ToDomain()

	return &deposit, nil
}

func (r DepositRepository) List(ctx context.Context, query domain.DepositQuery) ([]domain.Deposit, int64, error) {
	ctx, span := r.tracer.Start(ctx, "DepositRepository.List")
	defer span.End()

	rows := make([]depositRow, 0)

	selectQuery := r.readDB.NewSelect().
		Model((*TransactionDB)(nil)).
		ColumnExpr("t.*").
		ColumnExpr(depositColumns).
		Join("JOIN deposits AS d ON d.transaction_id = t.id").
		Where("t.merchant_id = ?", query.MerchantID).
		Where("t.type = ?", domain.TransactionTypeDeposit)

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

	deposits := make([]domain.Deposit, 0, len(rows))
	for i := range rows {
		deposits = append(deposits, rows[i].ToDomain())
	}

	return deposits, int64(totalItem), nil
}
