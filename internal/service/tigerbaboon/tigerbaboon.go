package tigerbaboon

import (
	"context"
	"errors"

	"github.com/mudflap-autobotz/payment-service-go-template/internal/domain"

	"github.com/mudflap-autobotz/payment-common/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type TigerbaboonService struct {
	tigerbaboonRepo domain.TigerbaboonRepository
	tracer          trace.Tracer
}

func NewTigerbaboonService(r domain.TigerbaboonRepository) *TigerbaboonService {
	return &TigerbaboonService{
		tigerbaboonRepo: r,
		tracer:          otel.Tracer("service.tigerbaboon"),
	}
}

func (s *TigerbaboonService) GetByID(ctx context.Context, id int) (*domain.Tigerbaboon, error) {
	ctx, span := s.tracer.Start(ctx, "TigerbaboonService.GetByID")
	defer span.End()

	tigerbaboon, err := s.tigerbaboonRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrTigerbaboonNotFound) {
			logger.Ctx(ctx).Debug().Int("tigerbaboon_id", id).Msg("tigerbaboon not found")
			return nil, err
		}

		logger.Ctx(ctx).Error().Int("tigerbaboon_id", id).Err(err).Msg("failed to get tigerbaboon by ID")
		return nil, err
	}

	return tigerbaboon, nil
}

func (s *TigerbaboonService) GetList(ctx context.Context, query domain.ListQuery) (*domain.TigerbaboonList, error) {
	ctx, span := s.tracer.Start(ctx, "TigerbaboonService.GetList")
	defer span.End()

	list, err := s.tigerbaboonRepo.GetList(ctx, query)
	if err != nil {
		logger.Ctx(ctx).Error().Err(err).Msg("failed to get tigerbaboon list")
		return nil, err
	}

	return list, nil
}

func (s *TigerbaboonService) Create(ctx context.Context, username string) (*domain.Tigerbaboon, error) {
	ctx, span := s.tracer.Start(ctx, "TigerbaboonService.Create")
	defer span.End()

	tigerbaboon, err := s.tigerbaboonRepo.Create(ctx, &domain.Tigerbaboon{Username: username})
	if err != nil {
		if errors.Is(err, domain.ErrTigerbaboonConflict) {
			logger.Ctx(ctx).Debug().Str("username", username).Msg("username already exists")
			return nil, err
		}

		logger.Ctx(ctx).Error().Str("username", username).Err(err).Msg("failed to create tigerbaboon")
		return nil, err
	}

	logger.Ctx(ctx).Info().Int("tigerbaboon_id", tigerbaboon.ID).Str("username", tigerbaboon.Username).Msg("tigerbaboon created")

	return tigerbaboon, nil
}

func (s *TigerbaboonService) Update(ctx context.Context, id int, username string) (*domain.Tigerbaboon, error) {
	ctx, span := s.tracer.Start(ctx, "TigerbaboonService.Update")
	defer span.End()

	tigerbaboon, err := s.tigerbaboonRepo.Update(ctx, &domain.Tigerbaboon{ID: id, Username: username})
	if err != nil {
		if errors.Is(err, domain.ErrTigerbaboonNotFound) || errors.Is(err, domain.ErrTigerbaboonConflict) {
			logger.Ctx(ctx).Debug().Int("tigerbaboon_id", id).Err(err).Msg("cannot update tigerbaboon")
			return nil, err
		}

		logger.Ctx(ctx).Error().Int("tigerbaboon_id", id).Err(err).Msg("failed to update tigerbaboon")
		return nil, err
	}

	logger.Ctx(ctx).Info().Int("tigerbaboon_id", tigerbaboon.ID).Msg("tigerbaboon updated")

	return tigerbaboon, nil
}

func (s *TigerbaboonService) Delete(ctx context.Context, id int) error {
	ctx, span := s.tracer.Start(ctx, "TigerbaboonService.Delete")
	defer span.End()

	if err := s.tigerbaboonRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrTigerbaboonNotFound) {
			logger.Ctx(ctx).Debug().Int("tigerbaboon_id", id).Msg("tigerbaboon not found")
			return err
		}

		logger.Ctx(ctx).Error().Int("tigerbaboon_id", id).Err(err).Msg("failed to delete tigerbaboon")
		return err
	}

	logger.Ctx(ctx).Info().Int("tigerbaboon_id", id).Msg("tigerbaboon deleted")

	return nil
}
