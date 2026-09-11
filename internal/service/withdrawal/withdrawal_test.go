package withdrawal_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain/mocks"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/service/withdrawal"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errDatabase = errors.New("database error")

func newConfig() *config.Config {
	return &config.Config{
		Kafka: config.KafkaConfig{
			TopicWithdrawalCreated: "withdrawal.created",
			PublishTimeout:         time.Second,
		},
		Transaction: config.TransactionConfig{
			MinWithdrawal:        "100.00",
			MaxWithdrawal:        "500000.00",
			SourceBankCodes:      []string{"KTB", "SCB", "BAY"},
			DestinationBankCodes: []string{"SCB", "KTB", "BAY", "BBL", "KBANK"},
		},
	}
}

func setup(t *testing.T) (*mocks.WithdrawalRepository, *mocks.EventPublisher, *withdrawal.WithdrawalService) {
	t.Helper()

	repo := mocks.NewWithdrawalRepository(t)
	publisher := mocks.NewEventPublisher(t)

	return repo, publisher, withdrawal.NewWithdrawalService(repo, publisher, newConfig())
}

func actor() domain.Actor {
	return domain.Actor{Email: "shop@example.com", IPAddress: "127.0.0.1", UserAgent: "curl"}
}

func validInput(merchantID uuid.UUID) domain.InitiateWithdrawal {
	return domain.InitiateWithdrawal{
		MerchantID:    merchantID,
		Amount:        "1000.00",
		BankCode:      "SCB",
		AccountNumber: "1234567890",
		AccountName:   "Shop Owner",
	}
}

func createdWithdrawal(merchantID uuid.UUID, amount string) *domain.Withdrawal {
	return &domain.Withdrawal{
		Transaction: domain.Transaction{
			ID:         uuid.New(),
			MerchantID: merchantID,
			Type:       domain.TransactionTypeWithdrawal,
			Amount:     amount,
			Status:     domain.TransactionStatusPending,
			CreatedAt:  time.Now().UTC(),
		},
		BankCode:      "SCB",
		AccountNumber: "1234567890",
		AccountName:   "Shop Owner",
	}
}

func TestInitiate(t *testing.T) {
	t.Run("creates a pending withdrawal", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()
		expected := createdWithdrawal(merchantID, "1000.00")

		repo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(input domain.CreateWithdrawal) bool {
				return input.MerchantID == merchantID &&
					input.Amount == "1000.00" &&
					input.BankCode == "SCB" &&
					input.AccountNumber == "1234567890" &&
					input.AccountName == "Shop Owner"
			}), mock.MatchedBy(func(entry *domain.AuditEntry) bool {
				return entry != nil && entry.Actor.UserID == uuid.Nil && entry.Actor.Email == "shop@example.com"
			})).
			Return(expected, nil).
			Once()

		publisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
			return event.Topic == "withdrawal.created" && event.Key == merchantID.String()
		})).Return(nil).Once()

		result, err := service.Initiate(context.Background(), actor(), validInput(merchantID))

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("normalizes the bank code and account number", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()

		repo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(input domain.CreateWithdrawal) bool {
				return input.BankCode == "SCB" && input.AccountNumber == "1234567890"
			}), mock.Anything).
			Return(createdWithdrawal(merchantID, "1000.00"), nil).
			Once()

		publisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()

		input := validInput(merchantID)
		input.BankCode = "  scb  "
		input.AccountNumber = "123-456 7890"

		_, err := service.Initiate(context.Background(), actor(), input)

		require.NoError(t, err)
	})

	t.Run("still succeeds when publishing the event fails", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()
		expected := createdWithdrawal(merchantID, "1000.00")

		repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(expected, nil).Once()
		publisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("broker unreachable")).Once()

		result, err := service.Initiate(context.Background(), actor(), validInput(merchantID))

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("rejects amounts outside the configured range", func(t *testing.T) {
		for _, amount := range []string{"99.99", "500000.01", "0", "-1", "abc", "100.001", ""} {
			repo, publisher, service := setup(t)
			input := validInput(uuid.New())
			input.Amount = amount

			result, err := service.Initiate(context.Background(), actor(), input)

			assert.ErrorIs(t, err, domain.ErrInvalidAmount, amount)
			assert.Nil(t, result, amount)
			repo.AssertNotCalled(t, "Create")
			publisher.AssertNotCalled(t, "Publish")
		}
	})

	t.Run("rejects unsupported bank codes", func(t *testing.T) {
		for _, bankCode := range []string{"MEGA", "", "  ", "scb1"} {
			repo, publisher, service := setup(t)
			input := validInput(uuid.New())
			input.BankCode = bankCode

			result, err := service.Initiate(context.Background(), actor(), input)

			assert.ErrorIs(t, err, domain.ErrInvalidBankCode, bankCode)
			assert.Nil(t, result, bankCode)
			repo.AssertNotCalled(t, "Create")
			publisher.AssertNotCalled(t, "Publish")
		}
	})

	t.Run("rejects malformed account numbers", func(t *testing.T) {
		for _, accountNumber := range []string{"123456789", "1234567890123456", "12345abcde", "", "----------"} {
			repo, publisher, service := setup(t)
			input := validInput(uuid.New())
			input.AccountNumber = accountNumber

			result, err := service.Initiate(context.Background(), actor(), input)

			assert.ErrorIs(t, err, domain.ErrInvalidAccountNumber, accountNumber)
			assert.Nil(t, result, accountNumber)
			repo.AssertNotCalled(t, "Create")
			publisher.AssertNotCalled(t, "Publish")
		}
	})

	t.Run("rejects an over-long account name", func(t *testing.T) {
		repo, publisher, service := setup(t)
		input := validInput(uuid.New())
		input.AccountName = strings.Repeat("x", 256)

		result, err := service.Initiate(context.Background(), actor(), input)

		require.Error(t, err)
		assert.Nil(t, result)
		repo.AssertNotCalled(t, "Create")
		publisher.AssertNotCalled(t, "Publish")
	})

	t.Run("rejects an over-long merchant reference", func(t *testing.T) {
		repo, publisher, service := setup(t)
		input := validInput(uuid.New())
		input.MerchantReference = strings.Repeat("x", domain.MaxMerchantReferenceLength+1)

		result, err := service.Initiate(context.Background(), actor(), input)

		require.Error(t, err)
		assert.Nil(t, result)
		repo.AssertNotCalled(t, "Create")
		publisher.AssertNotCalled(t, "Publish")
	})

	t.Run("propagates domain rejections from the repository", func(t *testing.T) {
		for _, expectedErr := range []error{
			domain.ErrInsufficientBalance,
			domain.ErrWalletNotFound,
			domain.ErrDuplicateReference,
			errDatabase,
		} {
			repo, publisher, service := setup(t)

			repo.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, expectedErr).Once()

			result, err := service.Initiate(context.Background(), actor(), validInput(uuid.New()))

			assert.ErrorIs(t, err, expectedErr)
			assert.Nil(t, result)
			publisher.AssertNotCalled(t, "Publish")
		}
	})

	t.Run("accepts a destination bank that is not a source bank", func(t *testing.T) {
		repo, publisher, service := setup(t)
		merchantID := uuid.New()

		repo.EXPECT().
			Create(mock.Anything, mock.MatchedBy(func(input domain.CreateWithdrawal) bool {
				return input.BankCode == "KBANK"
			}), mock.Anything).
			Return(createdWithdrawal(merchantID, "1000.00"), nil).
			Once()

		publisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()

		input := validInput(merchantID)
		input.BankCode = "KBANK"

		result, err := service.Initiate(context.Background(), actor(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("rejects a destination bank outside the list", func(t *testing.T) {
		repo, publisher, service := setup(t)
		input := validInput(uuid.New())
		input.BankCode = "MEGA"

		result, err := service.Initiate(context.Background(), actor(), input)

		assert.ErrorIs(t, err, domain.ErrInvalidBankCode)
		assert.Nil(t, result)
		repo.AssertNotCalled(t, "Create")
		publisher.AssertNotCalled(t, "Publish")
	})
}

func TestGet(t *testing.T) {
	t.Run("returns the withdrawal scoped to the merchant", func(t *testing.T) {
		repo, _, service := setup(t)
		merchantID := uuid.New()
		expected := createdWithdrawal(merchantID, "1000.00")

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
	t.Run("returns withdrawals with the total item count", func(t *testing.T) {
		repo, _, service := setup(t)
		merchantID := uuid.New()
		query := domain.WithdrawalQuery{
			ListQuery:  domain.ListQuery{Page: 1, Size: 10, SortBy: "created_at", OrderBy: "desc"},
			MerchantID: merchantID,
			Status:     domain.TransactionStatusPending,
		}
		withdrawals := []domain.Withdrawal{*createdWithdrawal(merchantID, "1000.00")}

		repo.EXPECT().List(mock.Anything, query).Return(withdrawals, int64(1), nil).Once()

		result, totalItem, err := service.List(context.Background(), query)

		require.NoError(t, err)
		assert.Equal(t, withdrawals, result)
		assert.Equal(t, int64(1), totalItem)
	})

	t.Run("propagates a repository failure", func(t *testing.T) {
		repo, _, service := setup(t)
		query := domain.WithdrawalQuery{MerchantID: uuid.New()}

		repo.EXPECT().List(mock.Anything, query).Return(nil, int64(0), errDatabase).Once()

		result, totalItem, err := service.List(context.Background(), query)

		assert.ErrorIs(t, err, errDatabase)
		assert.Nil(t, result)
		assert.Zero(t, totalItem)
	})
}
