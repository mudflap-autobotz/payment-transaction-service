package transaction

import (
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/uptrace/bun"
)

func applyTransactionFilters(query *bun.SelectQuery, list domain.ListQuery, dateFrom, dateTo *time.Time) *bun.SelectQuery {
	if list.Search != "" {
		pattern := "%" + list.Search + "%"
		query = query.Where("(t.id::text ILIKE ? OR t.merchant_reference ILIKE ? OR t.bank_reference ILIKE ?)", pattern, pattern, pattern)
	}

	if dateFrom != nil {
		query = query.Where("t.created_at >= ?", *dateFrom)
	}

	if dateTo != nil {
		query = query.Where("t.created_at <= ?", *dateTo)
	}

	return query
}
