package pgerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mudflap-autobotz/payment-transaction-service/internal/repository/pgerr"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func pgError(code string) error {
	return &pgconn.PgError{Code: code}
}

func TestClassifiers(t *testing.T) {
	tests := []struct {
		name       string
		classifier func(error) bool
		matching   error
	}{
		{"unique violation", pgerr.IsUniqueViolation, pgError(pgerrcode.UniqueViolation)},
		{"foreign key violation", pgerr.IsForeignKeyViolation, pgError(pgerrcode.ForeignKeyViolation)},
		{"check violation", pgerr.IsCheckViolation, pgError(pgerrcode.CheckViolation)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.True(t, test.classifier(test.matching))
			assert.True(t, test.classifier(fmt.Errorf("wrapped: %w", test.matching)))

			assert.False(t, test.classifier(nil))
			assert.False(t, test.classifier(errors.New("plain error")))
			assert.False(t, test.classifier(pgError(pgerrcode.SerializationFailure)))
		})
	}
}

func TestClassifiersDoNotOverlap(t *testing.T) {
	assert.False(t, pgerr.IsUniqueViolation(pgError(pgerrcode.CheckViolation)))
	assert.False(t, pgerr.IsCheckViolation(pgError(pgerrcode.UniqueViolation)))
	assert.False(t, pgerr.IsForeignKeyViolation(pgError(pgerrcode.UniqueViolation)))
}
