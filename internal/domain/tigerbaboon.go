package domain

import "context"

var (
	ErrTigerbaboonNotFound = NotFoundError("tigerbaboon")
	ErrTigerbaboonConflict = ConflictError("username")
)

type Tigerbaboon struct {
	ID       int
	Username string
}

type TigerbaboonList struct {
	Tigerbaboons []Tigerbaboon
	Total        int64
}

type TigerbaboonRepository interface {
	GetByID(ctx context.Context, id int) (*Tigerbaboon, error)
	GetList(ctx context.Context, query ListQuery) (*TigerbaboonList, error)
	Create(ctx context.Context, tigerbaboon *Tigerbaboon) (*Tigerbaboon, error)
	Update(ctx context.Context, tigerbaboon *Tigerbaboon) (*Tigerbaboon, error)
	Delete(ctx context.Context, id int) error
}

type TigerbaboonService interface {
	GetByID(ctx context.Context, id int) (*Tigerbaboon, error)
	GetList(ctx context.Context, query ListQuery) (*TigerbaboonList, error)
	Create(ctx context.Context, username string) (*Tigerbaboon, error)
	Update(ctx context.Context, id int, username string) (*Tigerbaboon, error)
	Delete(ctx context.Context, id int) error
}
