package domain

const (
	DefaultPaginationPage    = 1
	DefaultPaginationSize    = 10
	DefaultPaginationOrderBy = "desc"
	DefaultPaginationSortBy  = "id"
)

type ListQuery struct {
	Page    int
	Size    int
	Search  string
	OrderBy string
	SortBy  string
}

func (q ListQuery) Offset() int {
	if q.Page < 1 {
		return 0
	}
	return (q.Page - 1) * q.Size
}
