package repository

import (
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/repository/tigerbaboon"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	NewWriteDB,
	NewReadDB,
	tigerbaboon.NewTigerbaboonRepository,

	wire.Bind(new(domain.TigerbaboonRepository), new(*tigerbaboon.TigerbaboonRepository)),
)
