package token

import (
	"github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"

	"github.com/google/wire"
)

var TokenSet = wire.NewSet(
	NewIssuer,
)

func NewIssuer(cfg *config.Config) (*jwt.Issuer, error) {
	return jwt.New(jwt.Config{
		Secret: cfg.JWT.Secret,
		TTL:    cfg.JWT.TTL,
	})
}
