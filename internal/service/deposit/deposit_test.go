package deposit_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain/mocks"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/service/deposit"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errDatabase = errors.New("database error")

func newConfig() *config.Config {
	return &config.Config{
		Kafka: config.KafkaConfig{
			TopicDepositCreated: "deposit.created",
			PublishTimeout:      time.Second,
		},
		PromptPay: config.PromptPayConfig{
			TargetType:   "MOBILE",
			Target:       "0812345678",
			MerchantName: "PAYMENT GATEWAY",
			MerchantCity: "BANGKOK",
		},
		Transaction: config.TransactionConfig{
			MinDeposit: "100.00",
			MaxDeposit: "1000000.00",
			QRExpiry:   30 * time.Minute,
		},
	}
}

func setup(t *testing.T) (*mocks.DepositRepository, *mocks.EventPublisher, *deposit.DepositService) {
	t.Helper()

	repo := mocks.NewDepositRepository(t)
	publisher := mocks.NewEventPublisher(t)

	return repo, publisher, deposit.NewDepositService(repo, publisher, newConfig())
}

func actor() domain.Actor {
	return domain.Actor{Email: "shop@example.com", IPAddress: "127.0.0.1", UserAgent: "curl"}
}

func createdDeposit(merchantID uuid.UUID, amount string) *domain.Deposit {
	return &domain.Deposit{
		Transaction: domain.Transaction{
			ID:         uuid.New(),
			MerchantID: merchantID,
			Type:       domain.TransactionTypeDeposit,
			Amount:     amount,
			Status:     domain.TransactionStatusPending,
			CreatedAt:  time.Now().UTC(),
		},
		QRCode:     "000201010212",
		QRType:     domain.QRTypePromptPay,
		QRExpiryAt: time.Now().UTC().Add(30 * time.Minute),
	}
}

func TestInitiate(t *testing.T) {
	t.Run("creates a pending deposit with a promptpay qr", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()
		expected := createdDeposit(merchantID, "500.00")

		repo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(input domain.CreateDeposit) bool {
				return input.MerchantID == merchantID &&
					input.Amount == "500.00" &&
					input.QRType == domain.QRTypePromptPay &&
					strings.HasPrefix(input.QRCode, "000201010212") &&
					input.QRExpiryAt.After(time.Now().UTC())
			}), mock.MatchedBy(func(entry *domain.AuditEntry) bool {
				return entry != nil &&
					entry.EntityType == domain.AuditEntityTransaction &&
					entry.Action == domain.AuditActionCreated &&
					entry.Actor.Email == "shop@example.com" &&
					entry.Actor.UserID == uuid.Nil
			})).
			Return(expected, nil).
			Once()

		publisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
			return event.Topic == "deposit.created" && event.Key == merchantID.String()
		})).Return(nil).Once()

		result, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
			MerchantID: merchantID,
			Amount:     "500",
		})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("uses the merchant reference as the qr reference when provided", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()

		repo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(input domain.CreateDeposit) bool {
				return input.MerchantReference == "ORDER-1" && strings.Contains(input.QRCode, "ORDER-1")
			}), mock.Anything).
			Return(createdDeposit(merchantID, "500.00"), nil).
			Once()

		publisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()

		_, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
			MerchantID:        merchantID,
			Amount:            "500.00",
			MerchantReference: "  ORDER-1  ",
		})

		require.NoError(t, err)
	})

	t.Run("still succeeds when publishing the event fails", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()
		expected := createdDeposit(merchantID, "500.00")

		repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(expected, nil).Once()
		publisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("broker unreachable")).Once()

		result, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
			MerchantID: merchantID,
			Amount:     "500.00",
		})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("rejects amounts outside the configured range without touching the repository", func(t *testing.T) {
		for _, amount := range []string{"99.99", "1000000.01", "0", "-1", "abc", "100.001", ""} {
			repo, publisher, service := setup(t)

			result, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
				MerchantID: uuid.New(),
				Amount:     amount,
			})

			assert.ErrorIs(t, err, domain.ErrInvalidAmount, amount)
			assert.Nil(t, result, amount)
			repo.AssertNotCalled(t, "Create")
			publisher.AssertNotCalled(t, "Publish")
		}
	})

	t.Run("rejects an over-long merchant reference", func(t *testing.T) {
		repo, publisher, service := setup(t)

		result, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
			MerchantID:        uuid.New(),
			Amount:            "500.00",
			MerchantReference: strings.Repeat("x", domain.MaxMerchantReferenceLength+1),
		})

		require.Error(t, err)
		assert.Nil(t, result)
		repo.AssertNotCalled(t, "Create")
		publisher.AssertNotCalled(t, "Publish")
	})

	t.Run("propagates a duplicate reference conflict", func(t *testing.T) {
		repo, publisher, service := setup(t)

		repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
			Return(nil, domain.ErrDuplicateReference).Once()

		result, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
			MerchantID:        uuid.New(),
			Amount:            "500.00",
			MerchantReference: "ORDER-1",
		})

		assert.ErrorIs(t, err, domain.ErrDuplicateReference)
		assert.Nil(t, result)
		publisher.AssertNotCalled(t, "Publish")
	})

	t.Run("propagates a repository failure", func(t *testing.T) {
		repo, publisher, service := setup(t)

		repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errDatabase).Once()

		result, err := service.Initiate(context.Background(), actor(), domain.InitiateDeposit{
			MerchantID: uuid.New(),
			Amount:     "500.00",
		})

		assert.ErrorIs(t, err, errDatabase)
		assert.Nil(t, result)
		publisher.AssertNotCalled(t, "Publish")
	})
}

func TestGet(t *testing.T) {
	t.Run("returns the deposit scoped to the merchant", func(t *testing.T) {
		repo, _, service := setup(t)
		merchantID := uuid.New()
		expected := createdDeposit(merchantID, "500.00")

		repo.EXPECT().GetByID(mock.Anything, merchantID, expected.ID).Return(expected, nil).Once()

		result, err := service.Get(context.Background(), merchantID, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("propagates not found", func(t *testing.T) {
		repo, _, service := setup(t)
		merchantID := uuid.New()
		id := uuid.New()

		repo.EXPECT().GetByID(mock.Anything, merchantID, id).Return(nil, domain.ErrTransactionNotFound).Once()

		result, err := service.Get(context.Background(), merchantID, id)

		assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
		assert.Nil(t, result)
	})

	t.Run("propagates a repository failure", func(t *testing.T) {
		repo, _, service := setup(t)
		merchantID := uuid.New()
		id := uuid.New()

		repo.EXPECT().GetByID(mock.Anything, merchantID, id).Return(nil, errDatabase).Once()

		_, err := service.Get(context.Background(), merchantID, id)

		assert.ErrorIs(t, err, errDatabase)
	})
}

func TestList(t *testing.T) {
	t.Run("returns deposits with the total item count", func(t *testing.T) {
		repo, _, service := setup(t)
		merchantID := uuid.New()
		query := domain.DepositQuery{
			ListQuery:  domain.ListQuery{Page: 1, Size: 10, SortBy: "created_at", OrderBy: "desc"},
			MerchantID: merchantID,
			Status:     domain.TransactionStatusPending,
		}
		deposits := []domain.Deposit{*createdDeposit(merchantID, "500.00")}

		repo.EXPECT().List(mock.Anything, query).Return(deposits, int64(1), nil).Once()

		result, totalItem, err := service.List(context.Background(), query)

		require.NoError(t, err)
		assert.Equal(t, deposits, result)
		assert.Equal(t, int64(1), totalItem)
	})

	t.Run("propagates a repository failure", func(t *testing.T) {
		repo, _, service := setup(t)
		query := domain.DepositQuery{MerchantID: uuid.New()}

		repo.EXPECT().List(mock.Anything, query).Return(nil, int64(0), errDatabase).Once()

		result, totalItem, err := service.List(context.Background(), query)

		assert.ErrorIs(t, err, errDatabase)
		assert.Nil(t, result)
		assert.Zero(t, totalItem)
	})
}
