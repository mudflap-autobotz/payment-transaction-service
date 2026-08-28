package dto_test

import (
	"testing"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/dto"

	"github.com/stretchr/testify/assert"
)

func TestPaginationQueryApplyDefaults(t *testing.T) {
	t.Run("fills every zero value", func(t *testing.T) {
		query := dto.PaginationQuery{}

		query.ApplyDefaults()

		assert.Equal(t, domain.DefaultPaginationPage, query.Page)
		assert.Equal(t, domain.DefaultPaginationSize, query.Size)
		assert.Equal(t, domain.DefaultPaginationOrderBy, query.OrderBy)
		assert.Equal(t, domain.DefaultPaginationSortBy, query.SortBy)
	})

	t.Run("keeps values the caller supplied", func(t *testing.T) {
		query := dto.PaginationQuery{Page: 4, Size: 50, Search: "shop", SortBy: "amount", OrderBy: "asc"}

		query.ApplyDefaults()

		assert.Equal(t, 4, query.Page)
		assert.Equal(t, 50, query.Size)
		assert.Equal(t, "amount", query.SortBy)
		assert.Equal(t, "asc", query.OrderBy)
	})
}

func TestPaginationQueryToListQuery(t *testing.T) {
	query := dto.PaginationQuery{Page: 2, Size: 25, Search: "shop", SortBy: "created_at", OrderBy: "desc"}

	converted := query.ToListQuery()

	assert.Equal(t, 2, converted.Page)
	assert.Equal(t, 25, converted.Size)
	assert.Equal(t, "shop", converted.Search)
	assert.Equal(t, "created_at", converted.SortBy)
	assert.Equal(t, "desc", converted.OrderBy)
}
