package service

import (
	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/service/tigerbaboon"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	tigerbaboon.NewTigerbaboonService,

	wire.Bind(new(domain.TigerbaboonService), new(*tigerbaboon.TigerbaboonService)),
)
