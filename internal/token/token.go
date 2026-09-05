package token

import (
	"github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/google/wire"
)

var TokenSet = wire.NewSet(
	NewVerifier,
)

func NewVerifier(cfg *config.Config) (*jwt.Verifier, error) {
	return jwt.NewVerifier(jwt.VerifierConfig{
		PublicKey: cfg.JWT.PublicKey,
	})
}
