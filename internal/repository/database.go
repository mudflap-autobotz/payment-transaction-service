package repository

import (
	"github.com/mudflap-autobotz/payment-common/database"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
)

func NewWriteDB(cfg *config.Config) (*database.WriteDB, func(), error) {
	return database.NewWriteDB(toDatabaseConfig(cfg.Database.Postgres.Write))
}

func NewReadDB(cfg *config.Config) (*database.ReadDB, func(), error) {
	return database.NewReadDB(toDatabaseConfig(cfg.Database.Postgres.Read))
}

func toDatabaseConfig(c config.PostgresConnConfig) database.PostgresConfig {
	return database.PostgresConfig{
		Host:     c.Host,
		Port:     c.Port,
		User:     c.User,
		Password: c.Password,
		Name:     c.Name,
		SSLMode:  c.SSLMode,
		Pool: database.PoolConfig{
			MaxOpenConns:    c.Pool.MaxOpenConns,
			MaxIdleConns:    c.Pool.MaxIdleConns,
			ConnMaxLifetime: c.Pool.ConnMaxLifetime,
			ConnMaxIdleTime: c.Pool.ConnMaxIdleTime,
		},
	}
}
