//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/handler"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/publisher/kafka"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/repository"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/router"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/service"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/token"
)

func InitializeApp(cfg *config.Config) (*fiber.App, func(), error) {
	wire.Build(
		repository.RepositorySet,
		kafka.PublisherSet,
		token.TokenSet,
		service.ServiceSet,
		handler.HandlerSet,
		middleware.MiddlewareSet,
		router.RouterSet,
	)

	return &fiber.App{}, nil, nil
}
