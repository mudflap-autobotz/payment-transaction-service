package service

import (
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/service/deposit"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/service/withdrawal"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	deposit.NewDepositService,
	withdrawal.NewWithdrawalService,

	wire.Bind(new(domain.DepositService), new(*deposit.DepositService)),
	wire.Bind(new(domain.WithdrawalService), new(*withdrawal.WithdrawalService)),
)
