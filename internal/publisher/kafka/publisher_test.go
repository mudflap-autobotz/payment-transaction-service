package kafka_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/publisher/kafka"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func kafkaConfig(enabled bool, brokers []string) *config.Config {
	return &config.Config{
		Kafka: config.KafkaConfig{
			Enabled:                enabled,
			Brokers:                brokers,
			TopicDepositCreated:    domain.EventDepositCreated,
			TopicWithdrawalCreated: domain.EventWithdrawalCreated,
			WriteTimeout:           500 * time.Millisecond,
			PublishTimeout:         500 * time.Millisecond,
			AllowAutoTopicCreation: true,
		},
	}
}

func depositEvent() domain.Event {
	transactionID := uuid.New()

	return domain.Event{
		Topic: domain.EventDepositCreated,
		Key:   transactionID.String(),
		Value: kafka.DepositCreatedEvent{
			TransactionID: transactionID.String(),
			MerchantID:    uuid.New().String(),
			Amount:        "500.00",
			Status:        domain.TransactionStatusPending,
			QRCode:        "00020101021229370016A000000677010111",
			QRExpiryAt:    time.Now().UTC().Add(30 * time.Minute),
			CreatedAt:     time.Now().UTC(),
		},
	}
}

func TestNewEventPublisherReturnsANoopWhenKafkaIsDisabled(t *testing.T) {
	publisher, cleanup, err := kafka.NewEventPublisher(kafkaConfig(false, nil))
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	defer cleanup()

	assert.IsType(t, kafka.NoopPublisher{}, publisher)
	assert.NoError(t, publisher.Publish(context.Background(), depositEvent()))
}

func TestNewEventPublisherReturnsAWriterBackedPublisherWhenEnabled(t *testing.T) {
	publisher, cleanup, err := kafka.NewEventPublisher(kafkaConfig(true, []string{"localhost:9092"}))
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	defer cleanup()

	assert.IsType(t, &kafka.Publisher{}, publisher)
}

func TestPublishFailsWhenTheBrokerIsUnreachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	unreachable := listener.Addr().String()
	require.NoError(t, listener.Close())

	publisher, cleanup, err := kafka.NewEventPublisher(kafkaConfig(true, []string{unreachable}))
	require.NoError(t, err)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	assert.Error(t, publisher.Publish(ctx, depositEvent()))
}

func TestEventPayloadsMarshalToTheDocumentedShape(t *testing.T) {
	t.Run("deposit created", func(t *testing.T) {
		payload, err := json.Marshal(kafka.DepositCreatedEvent{
			TransactionID: "11111111-1111-1111-1111-111111111111",
			MerchantID:    "22222222-2222-2222-2222-222222222222",
			Amount:        "500.00",
			Status:        domain.TransactionStatusPending,
			QRCode:        "000201",
			QRExpiryAt:    time.Date(2026, 8, 28, 10, 30, 0, 0, time.UTC),
			CreatedAt:     time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(payload, &decoded))

		assert.Equal(t, "500.00", decoded["amount"])
		assert.Equal(t, domain.TransactionStatusPending, decoded["status"])
		assert.Equal(t, "2026-08-28T10:30:00Z", decoded["qr_expiry_at"])
		assert.NotContains(t, decoded, "merchant_reference")
	})

	t.Run("withdrawal created", func(t *testing.T) {
		payload, err := json.Marshal(kafka.WithdrawalCreatedEvent{
			TransactionID: "11111111-1111-1111-1111-111111111111",
			MerchantID:    "22222222-2222-2222-2222-222222222222",
			Amount:        "1000.00",
			Status:        domain.TransactionStatusPending,
			BankCode:      "SCB",
			AccountNumber: "1234567890",
			CreatedAt:     time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC),
		})
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(payload, &decoded))

		assert.Equal(t, "SCB", decoded["bank_code"])
		assert.Equal(t, "1234567890", decoded["account_number"])
		assert.NotContains(t, decoded, "account_name")
	})
}
