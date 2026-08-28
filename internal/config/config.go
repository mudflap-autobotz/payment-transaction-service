package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      AppConfig    `env-prefix:"APP_"`
	Database DBConfig     `env-prefix:"DB_"`
	Log      LogConfig    `env-prefix:"LOG_"`
	Tracer   TracerConfig `env-prefix:"TRACER_"`
}

type AppConfig struct {
	Name        string          `env:"NAME" env-required:"true"`
	Port        int             `env:"PORT" env-required:"true"`
	Environment string          `env:"ENVIRONMENT" env-required:"true"`
	RateLimit   RateLimitConfig `env-prefix:"RATE_LIMIT_"`
	Cors        CorsConfig      `env-prefix:"CORS_"`
}

type RateLimitConfig struct {
	MaxRequests int           `env:"MAX_REQUESTS" env-default:"100"`
	Expiration  time.Duration `env:"EXPIRATION" env-default:"1m"`
}

type CorsConfig struct {
	AllowOrigins     []string `env:"ALLOW_ORIGINS" env-default:"*"`
	AllowMethods     []string `env:"ALLOW_METHODS" env-default:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	AllowHeaders     []string `env:"ALLOW_HEADERS" env-default:"*"`
	AllowCredentials bool     `env:"ALLOW_CREDENTIALS" env-default:"false"`
}

type DBConfig struct {
	Postgres PostgresConfig `env-prefix:"POSTGRES_"`
	Mongo    MongoConfig    `env-prefix:"MONGO_"`
}

type PostgresConfig struct {
	Write PostgresConnConfig `env-prefix:"WRITE_"`
	Read  PostgresConnConfig `env-prefix:"READ_"`
}

type PostgresConnConfig struct {
	Host     string `env:"HOST" env-required:"true"`
	Port     int    `env:"PORT" env-required:"true"`
	User     string `env:"USER" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
	Name     string `env:"NAME" env-required:"true"`
	SSLMode  string `env:"SSL_MODE" env-required:"true"`
}

type MongoConfig struct {
	URI      string `env:"URI"`
	Database string `env:"DATABASE"`
}

type LogConfig struct {
	Level  string `env:"LEVEL" env-default:"info"`
	Format string `env:"FORMAT" env-default:"json"`
}

type TracerConfig struct {
	Enabled bool `env:"ENABLED" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			return nil, fmt.Errorf("Load config: %w", err)
		}
		return &cfg, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("Load config: %w", err)
	}
	return &cfg, nil
}
