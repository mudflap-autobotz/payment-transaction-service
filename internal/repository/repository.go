package repository

import (
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/repository/transaction"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	NewWriteDB,
	NewReadDB,
	transaction.NewDepositRepository,
	transaction.NewWithdrawalRepository,

	wire.Bind(new(domain.DepositRepository), new(*transaction.DepositRepository)),
	wire.Bind(new(domain.WithdrawalRepository), new(*transaction.WithdrawalRepository)),
)
