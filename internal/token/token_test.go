package token_test

import (
	"context"
	"testing"
	"time"

	commonjwt "github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-common/jwt/jwttest"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/token"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerifier(t *testing.T) {
	t.Run("verifies a token signed by the matching private key", func(t *testing.T) {
		keys := jwttest.MustGenerateKeys()

		verifier, err := token.NewVerifier(&config.Config{
			JWT: config.JWTConfig{PublicKey: keys.PublicPEM},
		})
		require.NoError(t, err)
		require.NotNil(t, verifier)

		signer, err := commonjwt.NewSigner(commonjwt.SignerConfig{
			PrivateKey: keys.PrivatePEM,
			TTL:        time.Hour,
		})
		require.NoError(t, err)

		merchantID := uuid.New()

		signed, err := signer.Issue(context.Background(), commonjwt.Claims{
			UserID: merchantID,
			Email:  "shop@example.com",
			Type:   domain.TokenTypeMerchant,
		})
		require.NoError(t, err)

		claims, err := verifier.Parse(signed)
		require.NoError(t, err)

		assert.Equal(t, merchantID, claims.UserID)
		assert.Equal(t, "shop@example.com", claims.Email)
		assert.Equal(t, domain.TokenTypeMerchant, claims.Type)
	})

	t.Run("rejects a token signed by another key pair", func(t *testing.T) {
		verifier, err := token.NewVerifier(&config.Config{
			JWT: config.JWTConfig{PublicKey: jwttest.MustGenerateKeys().PublicPEM},
		})
		require.NoError(t, err)

		signer, err := commonjwt.NewSigner(commonjwt.SignerConfig{
			PrivateKey: jwttest.MustGenerateKeys().PrivatePEM,
			TTL:        time.Hour,
		})
		require.NoError(t, err)

		signed, err := signer.Issue(context.Background(), commonjwt.Claims{
			UserID: uuid.New(),
			Type:   domain.TokenTypeMerchant,
		})
		require.NoError(t, err)

		claims, err := verifier.Parse(signed)

		require.ErrorIs(t, err, commonjwt.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("fails when the public key is missing", func(t *testing.T) {
		verifier, err := token.NewVerifier(&config.Config{})

		require.ErrorIs(t, err, commonjwt.ErrMissingPublicKey)
		assert.Nil(t, verifier)
	})
}
