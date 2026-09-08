package middleware

import (
	"github.com/mudflap-autobotz/payment-common/ratelimit"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/gofiber/fiber/v3"
)

func NewRatelimitStorage(cfg *config.Config) (fiber.Storage, func(), error) {
	return ratelimit.New(ratelimit.Config{
		Enabled:          cfg.Redis.Enabled,
		Addr:             cfg.Redis.Addr,
		Password:         cfg.Redis.Password,
		Database:         cfg.Redis.Database,
		PoolSize:         cfg.Redis.PoolSize,
		DialTimeout:      cfg.Redis.DialTimeout,
		ReadTimeout:      cfg.Redis.ReadTimeout,
		WriteTimeout:     cfg.Redis.WriteTimeout,
		FailoverCooldown: cfg.Redis.FailoverCooldown,
	})
}
