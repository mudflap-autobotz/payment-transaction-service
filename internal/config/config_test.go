package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requiredEnv() map[string]string {
	return map[string]string{
		"APP_NAME":                   "payment-transaction-service",
		"APP_PORT":                   "3001",
		"APP_ENVIRONMENT":            "test",
		"DB_POSTGRES_WRITE_HOST":     "localhost",
		"DB_POSTGRES_WRITE_PORT":     "5432",
		"DB_POSTGRES_WRITE_USER":     "postgres",
		"DB_POSTGRES_WRITE_PASSWORD": "postgres",
		"DB_POSTGRES_WRITE_NAME":     "payment_gateway",
		"DB_POSTGRES_WRITE_SSL_MODE": "disable",
		"DB_POSTGRES_READ_HOST":      "localhost",
		"DB_POSTGRES_READ_PORT":      "5432",
		"DB_POSTGRES_READ_USER":      "postgres",
		"DB_POSTGRES_READ_PASSWORD":  "postgres",
		"DB_POSTGRES_READ_NAME":      "payment_gateway",
		"DB_POSTGRES_READ_SSL_MODE":  "disable",
		"TRACER_ENABLED":             "false",
		"JWT_PUBLIC_KEY":             "test-public-key",
		"PROMPTPAY_TARGET":           "0812345678",
	}
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()

	for key, value := range env {
		t.Setenv(key, value)
	}
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	setEnv(t, requiredEnv())

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "payment-transaction-service", cfg.App.Name)
	assert.Equal(t, 3001, cfg.App.Port)

	assert.Equal(t, "test-public-key", cfg.JWT.PublicKey)

	assert.False(t, cfg.Kafka.Enabled)
	assert.Equal(t, []string{"localhost:9092"}, cfg.Kafka.Brokers)
	assert.Equal(t, "deposit.created", cfg.Kafka.TopicDepositCreated)
	assert.Equal(t, "withdrawal.created", cfg.Kafka.TopicWithdrawalCreated)
	assert.Equal(t, 5*time.Second, cfg.Kafka.WriteTimeout)
	assert.Equal(t, 5*time.Second, cfg.Kafka.PublishTimeout)
	assert.True(t, cfg.Kafka.AllowAutoTopicCreation)

	assert.Equal(t, "MOBILE", cfg.PromptPay.TargetType)
	assert.Equal(t, "0812345678", cfg.PromptPay.Target)
	assert.Equal(t, "PAYMENT GATEWAY", cfg.PromptPay.MerchantName)
	assert.Equal(t, "BANGKOK", cfg.PromptPay.MerchantCity)

	assert.Equal(t, "100.00", cfg.Transaction.MinDeposit)
	assert.Equal(t, "1000000.00", cfg.Transaction.MaxDeposit)
	assert.Equal(t, "100.00", cfg.Transaction.MinWithdrawal)
	assert.Equal(t, "500000.00", cfg.Transaction.MaxWithdrawal)
	assert.Equal(t, 30*time.Minute, cfg.Transaction.QRExpiry)
	assert.Equal(t, []string{"KTB", "SCB", "BAY"}, cfg.Transaction.SourceBankCodes)
	assert.Equal(t, []string{
		"KBANK", "SCB", "BBL", "KTB", "BAY", "TTB", "TMB", "GSB", "BAAC", "UOB", "GHB", "CIMB",
		"LNH", "KKB", "KKP", "KNK", "CITI", "SCBT", "TISCO", "ISBT", "HSBC", "ICBC", "TCRB",
	}, cfg.Transaction.DestinationBankCodes)
	assert.Len(t, cfg.Transaction.DestinationBankCodes, 23)

	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
	assert.Equal(t, 100, cfg.App.RateLimit.MaxRequests)
	assert.Equal(t, time.Minute, cfg.App.RateLimit.Expiration)
}

func TestLoadConfigOverridesDefaults(t *testing.T) {
	env := requiredEnv()
	env["KAFKA_ENABLED"] = "true"
	env["KAFKA_BROKERS"] = "kafka-a:9092,kafka-b:9092"
	env["PROMPTPAY_TARGET_TYPE"] = "NATID"
	env["PROMPTPAY_TARGET"] = "1234567890123"
	env["TRANSACTION_QR_EXPIRY"] = "10m"
	env["TRANSACTION_MIN_WITHDRAWAL"] = "1000.00"
	env["TRANSACTION_SOURCE_BANK_CODES"] = "KTB"
	env["TRANSACTION_DESTINATION_BANK_CODES"] = "SCB,KBANK"
	setEnv(t, env)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	assert.True(t, cfg.Kafka.Enabled)
	assert.Equal(t, []string{"kafka-a:9092", "kafka-b:9092"}, cfg.Kafka.Brokers)
	assert.Equal(t, "NATID", cfg.PromptPay.TargetType)
	assert.Equal(t, "1234567890123", cfg.PromptPay.Target)
	assert.Equal(t, 10*time.Minute, cfg.Transaction.QRExpiry)
	assert.Equal(t, "1000.00", cfg.Transaction.MinWithdrawal)
	assert.Equal(t, []string{"KTB"}, cfg.Transaction.SourceBankCodes)
	assert.Equal(t, []string{"SCB", "KBANK"}, cfg.Transaction.DestinationBankCodes)
}

func TestLoadConfigFailsWithoutRequiredEnv(t *testing.T) {
	tests := []string{
		"JWT_PUBLIC_KEY",
		"PROMPTPAY_TARGET",
		"APP_NAME",
		"DB_POSTGRES_WRITE_HOST",
		"TRACER_ENABLED",
	}

	for _, missing := range tests {
		t.Run("missing "+missing, func(t *testing.T) {
			setEnv(t, requiredEnv())
			require.NoError(t, os.Unsetenv(missing))

			_, err := config.LoadConfig()

			assert.Error(t, err)
		})
	}
}
