package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	logger "github.com/mudflap-autobotz/payment-common/logger"
	"github.com/mudflap-autobotz/payment-common/tracer"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// @title Payment Transaction Service API
// @version 1.0
// @description API documentation for the payment transaction service.
// @BasePath /api/v1
// @tag.name merchants/deposits
// @tag.x-displayName Deposits
// @tag.description Merchant deposits via PromptPay QR
// @tag.name merchants/withdrawals
// @tag.x-displayName Withdrawals
// @tag.description Merchant withdrawals via bank transfer
// @x-tagGroups [{"name":"Merchant","tags":["merchants/deposits","merchants/withdrawals"]}]
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token issued by the auth service
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	setupLogger(cfg)

	setupTracing(cfg)

	log.Info().Msgf("Starting %s on port %d in %s environment", cfg.App.Name, cfg.App.Port, cfg.App.Environment)

	app, cleanup, err := InitializeApp(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize application")
	}
	defer cleanup()

	go func() {
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		log.Info().Msg("Server listening on " + addr)
		if err := app.Listen(addr); err != nil {
			log.Fatal().Msg("Failed to start server: " + err.Error())
		}
	}()

	gracefullyStopServer(app)
}

func setupLogger(cfg *config.Config) {
	logger.InitLogger(cfg.Log.Level, cfg.Log.Format)
}

func setupTracing(cfg *config.Config) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if !cfg.Tracer.Enabled {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tracer.InitTracer(ctx, cfg.App.Name, cfg.App.Environment); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize tracer")
	}

}

func gracefullyStopServer(app *fiber.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Info().Msg("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Error during shutdown")
	}

	log.Info().Msg("Server stopped")
}
