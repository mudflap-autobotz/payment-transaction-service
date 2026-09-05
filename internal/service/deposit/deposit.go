package deposit

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mudflap-autobotz/payment-common/logger"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/publisher/kafka"
	"github.com/mudflap-autobotz/payment-transaction-service/pkg/utils/money"
	"github.com/mudflap-autobotz/payment-transaction-service/pkg/utils/promptpay"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type DepositService struct {
	depositRepo    domain.DepositRepository
	publisher      domain.EventPublisher
	promptPay      promptpay.Target
	limits         config.TransactionConfig
	topic          string
	publishTimeout time.Duration
	tracer         trace.Tracer
}

func NewDepositService(repo domain.DepositRepository, publisher domain.EventPublisher, cfg *config.Config) *DepositService {
	return &DepositService{
		depositRepo: repo,
		publisher:   publisher,
		promptPay: promptpay.Target{
			Type:  cfg.PromptPay.TargetType,
			Value: cfg.PromptPay.Target,
			Name:  cfg.PromptPay.MerchantName,
			City:  cfg.PromptPay.MerchantCity,
		},
		limits:         cfg.Transaction,
		topic:          cfg.Kafka.TopicDepositCreated,
		publishTimeout: cfg.Kafka.PublishTimeout,
		tracer:         otel.Tracer("service.deposit"),
	}
}

func (s *DepositService) Initiate(ctx context.Context, actor domain.Actor, input domain.InitiateDeposit) (*domain.Deposit, error) {
	ctx, span := s.tracer.Start(ctx, "deposit.initiate")
	defer span.End()

	span.SetAttributes(
		attribute.String("merchant_id", input.MerchantID.String()),
		attribute.String("amount", input.Amount),
	)

	amount, err := s.normalizeAmount(input.Amount)
	if err != nil {
		logger.Ctx(ctx).Debug().Str("amount", input.Amount).Msg("deposit amount rejected")

		return nil, err
	}

	merchantReference, err := normalizeMerchantReference(input.MerchantReference)
	if err != nil {
		return nil, err
	}

	reference := merchantReference
	if reference == "" {
		reference = uuid.New().String()
	}

	qrCode, err := promptpay.Generate(s.promptPay, amount, reference)
	if err != nil {
		logger.Ctx(ctx).Error().Err(err).Msg("failed to generate promptpay qr")

		return nil, err
	}

	auditEntry := domain.NewAuditEntry(&actor, domain.AuditEntityTransaction, domain.AuditActionCreated, uuid.Nil, nil,
		map[string]any{
			"type":   domain.TransactionTypeDeposit,
			"amount": amount,
			"status": domain.TransactionStatusPending,
		})

	created, err := s.depositRepo.Create(ctx, domain.CreateDeposit{
		MerchantID:        input.MerchantID,
		Amount:            amount,
		MerchantReference: merchantReference,
		QRCode:            qrCode,
		QRType:            domain.QRTypePromptPay,
		QRExpiryAt:        time.Now().UTC().Add(s.limits.QRExpiry),
	}, auditEntry)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateReference) {
			logger.Ctx(ctx).Debug().Str("merchant_reference", merchantReference).Msg("duplicate merchant reference")

			return nil, err
		}

		logger.Ctx(ctx).Error().Err(err).Str("merchant_id", input.MerchantID.String()).Msg("failed to create deposit")

		return nil, err
	}

	span.SetAttributes(attribute.String("transaction_id", created.ID.String()))

	s.publishCreated(ctx, created)

	logger.Ctx(ctx).Info().
		Str("transaction_id", created.ID.String()).
		Str("merchant_id", created.MerchantID.String()).
		Str("amount", created.Amount).
		Msg("deposit initiated")

	return created, nil
}

func (s *DepositService) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.Deposit, error) {
	ctx, span := s.tracer.Start(ctx, "deposit.fetched")
	defer span.End()

	span.SetAttributes(attribute.String("transaction_id", id.String()))

	deposit, err := s.depositRepo.GetByID(ctx, merchantID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			logger.Ctx(ctx).Debug().Str("transaction_id", id.String()).Msg("deposit not found")

			return nil, err
		}

		logger.Ctx(ctx).Error().Err(err).Str("transaction_id", id.String()).Msg("failed to get deposit")

		return nil, err
	}

	span.SetAttributes(attribute.String("status", deposit.Status))

	return deposit, nil
}

func (s *DepositService) List(ctx context.Context, query domain.DepositQuery) ([]domain.Deposit, int64, error) {
	ctx, span := s.tracer.Start(ctx, "deposits.listed")
	defer span.End()

	span.SetAttributes(attribute.String("status", query.Status))

	deposits, totalItem, err := s.depositRepo.List(ctx, query)
	if err != nil {
		logger.Ctx(ctx).Error().Err(err).Str("merchant_id", query.MerchantID.String()).Msg("failed to list deposits")

		return nil, 0, err
	}

	span.SetAttributes(attribute.Int64("count", totalItem))

	return deposits, totalItem, nil
}

func (s *DepositService) publishCreated(ctx context.Context, created *domain.Deposit) {
	publishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.publishTimeout)
	defer cancel()

	event := domain.Event{
		Topic: s.topic,
		Key:   created.MerchantID.String(),
		Value: kafka.DepositCreatedEvent{
			TransactionID:     created.ID.String(),
			MerchantID:        created.MerchantID.String(),
			Amount:            created.Amount,
			Status:            created.Status,
			QRCode:            created.QRCode,
			QRExpiryAt:        created.QRExpiryAt,
			MerchantReference: created.MerchantReference,
			CreatedAt:         created.CreatedAt,
		},
	}

	if err := s.publisher.Publish(publishCtx, event); err != nil {
		trace.SpanFromContext(ctx).RecordError(err)

		logger.Ctx(ctx).Error().Err(err).
			Str("event", domain.EventDepositCreated).
			Str("transaction_id", created.ID.String()).
			Msg("failed to publish deposit created event")
	}
}

func (s *DepositService) normalizeAmount(raw string) (string, error) {
	amount, err := money.Normalize(raw)
	if err != nil {
		return "", domain.ErrInvalidAmount
	}

	withinLimits, err := money.InRange(amount, s.limits.MinDeposit, s.limits.MaxDeposit)
	if err != nil || !withinLimits {
		return "", domain.ErrInvalidAmount
	}

	return amount, nil
}

func normalizeMerchantReference(raw string) (string, error) {
	normalized := strings.TrimSpace(raw)
	if len(normalized) > domain.MaxMerchantReferenceLength {
		return "", domain.InvalidInputError("merchant_reference is too long")
	}

	return normalized, nil
}
