package tigerbaboon

import (
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"

	"github.com/uptrace/bun"
)

type TigerbaboonDB struct {
	bun.BaseModel `bun:"table:tigerbaboons,alias:t"`

	ID       int    `bun:"id,pk,autoincrement"`
	Username string `bun:"username"`
}

func (m *TigerbaboonDB) toDomain() *domain.Tigerbaboon {
	return &domain.Tigerbaboon{
		ID:       m.ID,
		Username: m.Username,
	}
}

func fromDomain(d *domain.Tigerbaboon) *TigerbaboonDB {
	return &TigerbaboonDB{
		ID:       d.ID,
		Username: d.Username,
	}
}
