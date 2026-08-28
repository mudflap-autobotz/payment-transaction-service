//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/wire"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/config"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/handler"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/middleware"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/repository"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/router"
	"github.com/mudflap-autobotz/payment-service-go-template/internal/service"
)

func InitializeApp(cfg *config.Config) (*fiber.App, func(), error) {
	wire.Build(
		repository.RepositorySet,
		service.ServiceSet,
		handler.HandlerSet,
		middleware.MiddlewareSet,
		router.RouterSet,
	)

	return &fiber.App{}, nil, nil
}
