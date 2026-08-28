package dto

import "github.com/mudflap-autobotz/payment-service-go-template/internal/domain"

type CreateTigerbaboonRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
}

type UpdateTigerbaboonRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
}

type GetTigerbaboonByIDRequest struct {
	ID int `uri:"id" validate:"required,min=1"`
}

type GetTigerbaboonListRequest struct {
	PaginationQuery
	SortBy string `query:"sort_by" validate:"required,oneof=id username"`
}

func (r *GetTigerbaboonListRequest) ApplyDefaults() {
	r.PaginationQuery.ApplyDefaults()

	if r.SortBy == "" {
		r.SortBy = domain.DefaultPaginationSortBy
	}
}

func (r GetTigerbaboonListRequest) ToListQuery() domain.ListQuery {
	return r.toListQuery(r.SortBy)
}

type TigerbaboonResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}
