package withdrawal

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

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	minAccountNumberLength = 10
	maxAccountNumberLength = 15
	maxAccountNameLength   = 255
)

type WithdrawalService struct {
	withdrawalRepo   domain.WithdrawalRepository
	publisher        domain.EventPublisher
	limits           config.TransactionConfig
	allowedBankCodes map[string]bool
	topic            string
	publishTimeout   time.Duration
	tracer           trace.Tracer
}

func NewWithdrawalService(repo domain.WithdrawalRepository, publisher domain.EventPublisher, cfg *config.Config) *WithdrawalService {
	allowedBankCodes := make(map[string]bool, len(cfg.Transaction.AllowedBankCodes))
	for _, bankCode := range cfg.Transaction.AllowedBankCodes {
		allowedBankCodes[strings.ToUpper(strings.TrimSpace(bankCode))] = true
	}

	return &WithdrawalService{
		withdrawalRepo:   repo,
		publisher:        publisher,
		limits:           cfg.Transaction,
		allowedBankCodes: allowedBankCodes,
		topic:            cfg.Kafka.TopicWithdrawalCreated,
		publishTimeout:   cfg.Kafka.PublishTimeout,
		tracer:           otel.Tracer("service.withdrawal"),
	}
}

func (s *WithdrawalService) Initiate(ctx context.Context, actor domain.Actor, input domain.InitiateWithdrawal) (*domain.Withdrawal, error) {
	ctx, span := s.tracer.Start(ctx, "withdrawal.initiate")
	defer span.End()

	span.SetAttributes(
		attribute.String("merchant_id", input.MerchantID.String()),
		attribute.String("amount", input.Amount),
		attribute.String("bank_code", input.BankCode),
	)

	amount, err := s.normalizeAmount(input.Amount)
	if err != nil {
		logger.Ctx(ctx).Debug().Str("amount", input.Amount).Msg("withdrawal amount rejected")

		return nil, err
	}

	bankCode, err := s.normalizeBankCode(input.BankCode)
	if err != nil {
		logger.Ctx(ctx).Debug().Str("bank_code", input.BankCode).Msg("withdrawal bank code rejected")

		return nil, err
	}

	accountNumber, err := normalizeAccountNumber(input.AccountNumber)
	if err != nil {
		logger.Ctx(ctx).Debug().Msg("withdrawal account number rejected")

		return nil, err
	}

	accountName, err := normalizeAccountName(input.AccountName)
	if err != nil {
		return nil, err
	}

	merchantReference, err := normalizeMerchantReference(input.MerchantReference)
	if err != nil {
		return nil, err
	}

	auditEntry := domain.NewAuditEntry(&actor, domain.AuditEntityTransaction, domain.AuditActionCreated, uuid.Nil, nil,
		map[string]any{
			"type":      domain.TransactionTypeWithdrawal,
			"amount":    amount,
			"status":    domain.TransactionStatusPending,
			"bank_code": bankCode,
		})

	created, err := s.withdrawalRepo.Create(ctx, domain.CreateWithdrawal{
		MerchantID:        input.MerchantID,
		Amount:            amount,
		MerchantReference: merchantReference,
		BankCode:          bankCode,
		AccountNumber:     accountNumber,
		AccountName:       accountName,
	}, auditEntry)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientBalance):
			logger.Ctx(ctx).Debug().Str("merchant_id", input.MerchantID.String()).Str("amount", amount).Msg("insufficient balance")
		case errors.Is(err, domain.ErrWalletNotFound), errors.Is(err, domain.ErrDuplicateReference):
			logger.Ctx(ctx).Debug().Str("merchant_id", input.MerchantID.String()).Err(err).Msg("withdrawal rejected")
		default:
			logger.Ctx(ctx).Error().Err(err).Str("merchant_id", input.MerchantID.String()).Msg("failed to create withdrawal")
		}

		return nil, err
	}

	span.SetAttributes(attribute.String("transaction_id", created.ID.String()))

	s.publishCreated(ctx, created)

	logger.Ctx(ctx).Info().
		Str("transaction_id", created.ID.String()).
		Str("merchant_id", created.MerchantID.String()).
		Str("amount", created.Amount).
		Str("bank_code", created.BankCode).
		Msg("withdrawal initiated")

	return created, nil
}

func (s *WithdrawalService) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.Withdrawal, error) {
	ctx, span := s.tracer.Start(ctx, "withdrawal.fetched")
	defer span.End()

	span.SetAttributes(attribute.String("transaction_id", id.String()))

	withdrawal, err := s.withdrawalRepo.GetByID(ctx, merchantID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			logger.Ctx(ctx).Debug().Str("transaction_id", id.String()).Msg("withdrawal not found")

			return nil, err
		}

		logger.Ctx(ctx).Error().Err(err).Str("transaction_id", id.String()).Msg("failed to get withdrawal")

		return nil, err
	}

	span.SetAttributes(attribute.String("status", withdrawal.Status))

	return withdrawal, nil
}

func (s *WithdrawalService) List(ctx context.Context, query domain.WithdrawalQuery) ([]domain.Withdrawal, int64, error) {
	ctx, span := s.tracer.Start(ctx, "withdrawals.listed")
	defer span.End()

	span.SetAttributes(attribute.String("status", query.Status))

	withdrawals, totalItem, err := s.withdrawalRepo.List(ctx, query)
	if err != nil {
		logger.Ctx(ctx).Error().Err(err).Str("merchant_id", query.MerchantID.String()).Msg("failed to list withdrawals")

		return nil, 0, err
	}

	span.SetAttributes(attribute.Int64("count", totalItem))

	return withdrawals, totalItem, nil
}

func (s *WithdrawalService) publishCreated(ctx context.Context, created *domain.Withdrawal) {
	publishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.publishTimeout)
	defer cancel()

	event := domain.Event{
		Topic: s.topic,
		Key:   created.MerchantID.String(),
		Value: kafka.WithdrawalCreatedEvent{
			TransactionID:     created.ID.String(),
			MerchantID:        created.MerchantID.String(),
			Amount:            created.Amount,
			Status:            created.Status,
			BankCode:          created.BankCode,
			AccountNumber:     created.AccountNumber,
			AccountName:       created.AccountName,
			MerchantReference: created.MerchantReference,
			CreatedAt:         created.CreatedAt,
		},
	}

	if err := s.publisher.Publish(publishCtx, event); err != nil {
		trace.SpanFromContext(ctx).RecordError(err)

		logger.Ctx(ctx).Error().Err(err).
			Str("event", domain.EventWithdrawalCreated).
			Str("transaction_id", created.ID.String()).
			Msg("failed to publish withdrawal created event")
	}
}

func (s *WithdrawalService) normalizeAmount(raw string) (string, error) {
	amount, err := money.Normalize(raw)
	if err != nil {
		return "", domain.ErrInvalidAmount
	}

	withinLimits, err := money.InRange(amount, s.limits.MinWithdrawal, s.limits.MaxWithdrawal)
	if err != nil || !withinLimits {
		return "", domain.ErrInvalidAmount
	}

	return amount, nil
}

func (s *WithdrawalService) normalizeBankCode(raw string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if !s.allowedBankCodes[normalized] {
		return "", domain.ErrInvalidBankCode
	}

	return normalized, nil
}

func normalizeAccountNumber(raw string) (string, error) {
	var digits strings.Builder

	for _, character := range strings.TrimSpace(raw) {
		switch {
		case character >= '0' && character <= '9':
			digits.WriteRune(character)
		case character == '-' || character == ' ':
		default:
			return "", domain.ErrInvalidAccountNumber
		}
	}

	normalized := digits.String()
	if len(normalized) < minAccountNumberLength || len(normalized) > maxAccountNumberLength {
		return "", domain.ErrInvalidAccountNumber
	}

	return normalized, nil
}

func normalizeAccountName(raw string) (string, error) {
	normalized := strings.TrimSpace(raw)
	if len(normalized) > maxAccountNameLength {
		return "", domain.InvalidInputError("account_name is too long")
	}

	return normalized, nil
}

func normalizeMerchantReference(raw string) (string, error) {
	normalized := strings.TrimSpace(raw)
	if len(normalized) > domain.MaxMerchantReferenceLength {
		return "", domain.InvalidInputError("merchant_reference is too long")
	}

	return normalized, nil
}
