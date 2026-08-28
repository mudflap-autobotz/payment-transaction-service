package token_test

import (
	"context"
	"testing"
	"time"

	commonjwt "github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/token"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIssuer(t *testing.T) {
	t.Run("builds an issuer that round-trips merchant claims", func(t *testing.T) {
		issuer, err := token.NewIssuer(&config.Config{
			JWT: config.JWTConfig{Secret: "transaction-service-token-secret", TTL: time.Hour},
		})
		require.NoError(t, err)
		require.NotNil(t, issuer)

		merchantID := uuid.New()

		signed, err := issuer.Issue(context.Background(), commonjwt.Claims{
			UserID: merchantID,
			Email:  "shop@example.com",
			Type:   domain.TokenTypeMerchant,
		})
		require.NoError(t, err)

		claims, err := issuer.Parse(signed)
		require.NoError(t, err)

		assert.Equal(t, merchantID, claims.UserID)
		assert.Equal(t, "shop@example.com", claims.Email)
		assert.Equal(t, domain.TokenTypeMerchant, claims.Type)
	})

	t.Run("fails when the secret is missing", func(t *testing.T) {
		issuer, err := token.NewIssuer(&config.Config{JWT: config.JWTConfig{TTL: time.Hour}})

		require.Error(t, err)
		assert.Nil(t, issuer)
	})
}
