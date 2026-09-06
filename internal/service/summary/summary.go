package summary

import (
	"context"
	"time"

	"github.com/mudflap-autobotz/payment-common/logger"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SummaryService struct {
	summaryRepo domain.SummaryRepository
	location    *time.Location
	tracer      trace.Tracer
}

func NewSummaryService(repo domain.SummaryRepository, cfg *config.Config) *SummaryService {
	return &SummaryService{
		summaryRepo: repo,
		location:    cfg.App.Location(),
		tracer:      otel.Tracer("service.summary"),
	}
}

func (s *SummaryService) Merchant(ctx context.Context, merchantID uuid.UUID) (*domain.MerchantSummary, error) {
	ctx, span := s.tracer.Start(ctx, "summary.merchant")
	defer span.End()

	span.SetAttributes(attribute.String("merchant_id", merchantID.String()))

	since := domain.StartOfDay(time.Now(), s.location)

	summary, err := s.summaryRepo.MerchantSummary(ctx, merchantID, since)
	if err != nil {
		logger.Ctx(ctx).Error().Err(err).Str("merchant_id", merchantID.String()).Msg("failed to build merchant summary")

		return nil, err
	}

	return summary, nil
}
