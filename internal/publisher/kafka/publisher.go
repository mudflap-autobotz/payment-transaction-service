package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/google/wire"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var PublisherSet = wire.NewSet(
	NewEventPublisher,
)

type Publisher struct {
	writer *kafka.Writer
	tracer trace.Tracer
}

type NoopPublisher struct{}

func NewEventPublisher(cfg *config.Config) (domain.EventPublisher, func(), error) {
	if !cfg.Kafka.Enabled {
		return NoopPublisher{}, func() {}, nil
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafka.Hash{},
		WriteTimeout:           cfg.Kafka.WriteTimeout,
		RequiredAcks:           kafka.RequireAll,
		Async:                  false,
		AllowAutoTopicCreation: cfg.Kafka.AllowAutoTopicCreation,
	}

	publisher := &Publisher{
		writer: writer,
		tracer: otel.Tracer("publisher.kafka"),
	}

	return publisher, func() { _ = writer.Close() }, nil
}

func (p *Publisher) Publish(ctx context.Context, event domain.Event) error {
	ctx, span := p.tracer.Start(ctx, "Publisher.Publish")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", event.Topic),
		attribute.String("messaging.message.key", event.Key),
	)

	body, err := json.Marshal(envelope{
		Event:      event.Topic,
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		Data:       event.Value,
	})
	if err != nil {
		span.RecordError(err)

		return err
	}

	message := kafka.Message{
		Topic:   event.Topic,
		Key:     []byte(event.Key),
		Value:   body,
		Headers: traceHeaders(ctx),
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		span.RecordError(err)

		return err
	}

	return nil
}

func (NoopPublisher) Publish(context.Context, domain.Event) error {
	return nil
}

func traceHeaders(ctx context.Context) []kafka.Header {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	headers := make([]kafka.Header, 0, len(carrier))
	for key, value := range carrier {
		headers = append(headers, kafka.Header{Key: key, Value: []byte(value)})
	}

	return headers
}
