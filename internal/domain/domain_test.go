package domain_test

import (
	"testing"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditEntry(t *testing.T) {
	t.Run("copies the actor and the change payload", func(t *testing.T) {
		actor := domain.Actor{Email: "shop@example.com", IPAddress: "203.0.113.9", UserAgent: "curl/8"}
		entityID := uuid.New()
		newValue := map[string]string{"status": domain.TransactionStatusPending}

		entry := domain.NewAuditEntry(&actor, domain.AuditEntityTransaction, domain.AuditActionCreated, entityID, nil, newValue)

		require.NotNil(t, entry)
		assert.Equal(t, domain.AuditEntityTransaction, entry.EntityType)
		assert.Equal(t, domain.AuditActionCreated, entry.Action)
		assert.Equal(t, entityID, entry.EntityID)
		assert.Nil(t, entry.OldValue)
		assert.Equal(t, newValue, entry.NewValue)
		assert.Equal(t, actor, entry.Actor)
	})

	t.Run("leaves the user id zero for a merchant actor", func(t *testing.T) {
		entry := domain.NewAuditEntry(
			&domain.Actor{Email: "shop@example.com"},
			domain.AuditEntityTransaction,
			domain.AuditActionCreated,
			uuid.New(),
			nil,
			nil,
		)

		require.NotNil(t, entry)
		assert.Equal(t, uuid.Nil, entry.Actor.UserID)
	})

	t.Run("returns nil without an actor", func(t *testing.T) {
		assert.Nil(t, domain.NewAuditEntry(nil, domain.AuditEntityTransaction, domain.AuditActionCreated, uuid.New(), nil, nil))
	})
}

func TestListQueryOffset(t *testing.T) {
	tests := []struct {
		name     string
		query    domain.ListQuery
		expected int
	}{
		{"first page", domain.ListQuery{Page: 1, Size: 10}, 0},
		{"third page", domain.ListQuery{Page: 3, Size: 10}, 20},
		{"large page size", domain.ListQuery{Page: 2, Size: 100}, 100},
		{"zero page falls back to no offset", domain.ListQuery{Page: 0, Size: 10}, 0},
		{"negative page falls back to no offset", domain.ListQuery{Page: -5, Size: 10}, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.query.Offset())
		})
	}
}
