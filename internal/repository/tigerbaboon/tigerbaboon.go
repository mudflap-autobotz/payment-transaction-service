package tigerbaboon

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/mudflap-autobotz/payment-common/database"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type TigerbaboonRepository struct {
	writeDB *database.WriteDB
	readDB  *database.ReadDB
	tracer  trace.Tracer
}

func NewTigerbaboonRepository(writeDB *database.WriteDB, readDB *database.ReadDB) *TigerbaboonRepository {
	return &TigerbaboonRepository{
		writeDB: writeDB,
		readDB:  readDB,
		tracer:  otel.Tracer("repository.tigerbaboon"),
	}
}

func (r TigerbaboonRepository) GetByID(ctx context.Context, id int) (*domain.Tigerbaboon, error) {
	ctx, span := r.tracer.Start(ctx, "TigerbaboonRepository.GetByID")
	defer span.End()

	tigerbaboonDB := new(TigerbaboonDB)
	if err := r.readDB.NewSelect().Model(tigerbaboonDB).Where("id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTigerbaboonNotFound
		}
		return nil, err
	}

	return tigerbaboonDB.toDomain(), nil
}

func (r TigerbaboonRepository) GetList(ctx context.Context, query domain.ListQuery) (*domain.TigerbaboonList, error) {
	ctx, span := r.tracer.Start(ctx, "TigerbaboonRepository.GetList")
	defer span.End()

	var tigerbaboonsDB []TigerbaboonDB

	q := r.readDB.NewSelect().Model(&tigerbaboonsDB)

	if query.Search != "" {
		q = q.Where("username ILIKE ?", "%"+query.Search+"%")
	}

	total, err := q.
		OrderExpr("? ?", bun.Ident(sortColumn(query.SortBy)), bun.Safe(sortDirection(query.OrderBy))).
		Limit(query.Size).
		Offset(query.Offset()).
		ScanAndCount(ctx)
	if err != nil {
		return nil, err
	}

	tigerbaboons := make([]domain.Tigerbaboon, 0, len(tigerbaboonsDB))
	for i := range tigerbaboonsDB {
		tigerbaboons = append(tigerbaboons, *tigerbaboonsDB[i].toDomain())
	}

	return &domain.TigerbaboonList{Tigerbaboons: tigerbaboons, Total: int64(total)}, nil
}

func (r TigerbaboonRepository) Create(ctx context.Context, tigerbaboon *domain.Tigerbaboon) (*domain.Tigerbaboon, error) {
	ctx, span := r.tracer.Start(ctx, "TigerbaboonRepository.Create")
	defer span.End()

	tigerbaboonDB := fromDomain(tigerbaboon)

	if _, err := r.writeDB.NewInsert().Model(tigerbaboonDB).Returning("*").Exec(ctx); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrTigerbaboonConflict
		}
		return nil, err
	}

	return tigerbaboonDB.toDomain(), nil
}

func (r TigerbaboonRepository) Update(ctx context.Context, tigerbaboon *domain.Tigerbaboon) (*domain.Tigerbaboon, error) {
	ctx, span := r.tracer.Start(ctx, "TigerbaboonRepository.Update")
	defer span.End()

	tigerbaboonDB := fromDomain(tigerbaboon)

	result, err := r.writeDB.NewUpdate().
		Model(tigerbaboonDB).
		Column("username").
		WherePK().
		Returning("*").
		Exec(ctx)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrTigerbaboonConflict
		}
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, domain.ErrTigerbaboonNotFound
	}

	return tigerbaboonDB.toDomain(), nil
}

func (r TigerbaboonRepository) Delete(ctx context.Context, id int) error {
	ctx, span := r.tracer.Start(ctx, "TigerbaboonRepository.Delete")
	defer span.End()

	result, err := r.writeDB.NewDelete().Model((*TigerbaboonDB)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrTigerbaboonNotFound
	}

	return nil
}

func (r TigerbaboonRepository) CreateBulk(ctx context.Context, usernames []string) (int64, error) {
	ctx, span := r.tracer.Start(ctx, "TigerbaboonRepository.CreateBulk")
	defer span.End()

	if len(usernames) == 0 {
		return 0, nil
	}

	rows := make([][]any, 0, len(usernames))
	for _, username := range usernames {
		rows = append(rows, []any{username})
	}

	var inserted int64
	err := r.writeDB.Pgx(ctx, func(conn *pgx.Conn) error {
		count, err := conn.CopyFrom(
			ctx,
			pgx.Identifier{"tigerbaboons"},
			[]string{"username"},
			pgx.CopyFromRows(rows),
		)
		inserted = count
		return err
	})
	if err != nil {
		if isUniqueViolation(err) {
			return 0, domain.ErrTigerbaboonConflict
		}
		return 0, err
	}

	return inserted, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

func sortColumn(sortBy string) string {
	switch sortBy {
	case "username":
		return "username"
	default:
		return "id"
	}
}

func sortDirection(orderBy string) string {
	if strings.EqualFold(orderBy, "asc") {
		return "ASC"
	}
	return "DESC"
}
