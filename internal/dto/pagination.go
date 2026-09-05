package dto

import "github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

type PaginationQuery struct {
	Page    int    `query:"page" validate:"required,min=1"`
	Size    int    `query:"size" validate:"required,min=1,max=100"`
	Search  string `query:"search"`
	SortBy  string `query:"sort_by" validate:"required,oneof=created_at updated_at amount status"`
	OrderBy string `query:"order_by" validate:"required,oneof=asc desc"`
}

func (q *PaginationQuery) ApplyDefaults() {
	if q.Page == 0 {
		q.Page = domain.DefaultPaginationPage
	}

	if q.Size == 0 {
		q.Size = domain.DefaultPaginationSize
	}

	if q.OrderBy == "" {
		q.OrderBy = domain.DefaultPaginationOrderBy
	}

	if q.SortBy == "" {
		q.SortBy = domain.DefaultPaginationSortBy
	}
}

func (q PaginationQuery) ToListQuery() domain.ListQuery {
	return domain.ListQuery{
		Page:    q.Page,
		Size:    q.Size,
		Search:  q.Search,
		OrderBy: q.OrderBy,
		SortBy:  q.SortBy,
	}
}
