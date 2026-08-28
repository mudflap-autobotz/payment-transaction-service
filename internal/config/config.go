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

	JWT         JWTConfig         `env-prefix:"JWT_"`
	Kafka       KafkaConfig       `env-prefix:"KAFKA_"`
	PromptPay   PromptPayConfig   `env-prefix:"PROMPTPAY_"`
	Transaction TransactionConfig `env-prefix:"TRANSACTION_"`
}

type JWTConfig struct {
	Secret string        `env:"SECRET" env-required:"true"`
	TTL    time.Duration `env:"TTL" env-default:"24h"`
}

type KafkaConfig struct {
	Enabled                bool          `env:"ENABLED" env-default:"false"`
	Brokers                []string      `env:"BROKERS" env-default:"localhost:9092"`
	TopicDepositCreated    string        `env:"TOPIC_DEPOSIT_CREATED" env-default:"deposit.created"`
	TopicWithdrawalCreated string        `env:"TOPIC_WITHDRAWAL_CREATED" env-default:"withdrawal.created"`
	WriteTimeout           time.Duration `env:"WRITE_TIMEOUT" env-default:"5s"`
	PublishTimeout         time.Duration `env:"PUBLISH_TIMEOUT" env-default:"5s"`
	AllowAutoTopicCreation bool          `env:"ALLOW_AUTO_TOPIC_CREATION" env-default:"true"`
}

type PromptPayConfig struct {
	TargetType   string `env:"TARGET_TYPE" env-default:"MOBILE"`
	Target       string `env:"TARGET" env-required:"true"`
	MerchantName string `env:"MERCHANT_NAME" env-default:"PAYMENT GATEWAY"`
	MerchantCity string `env:"MERCHANT_CITY" env-default:"BANGKOK"`
}

type TransactionConfig struct {
	MinDeposit       string        `env:"MIN_DEPOSIT" env-default:"100.00"`
	MaxDeposit       string        `env:"MAX_DEPOSIT" env-default:"1000000.00"`
	MinWithdrawal    string        `env:"MIN_WITHDRAWAL" env-default:"100.00"`
	MaxWithdrawal    string        `env:"MAX_WITHDRAWAL" env-default:"500000.00"`
	QRExpiry         time.Duration `env:"QR_EXPIRY" env-default:"30m"`
	AllowedBankCodes []string      `env:"ALLOWED_BANK_CODES" env-default:"SCB,KTB,BAY,BBL"`
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
