package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/middleware"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/token"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecret    = "middleware-transaction-unit-test-secret"
	merchantEmail = "shop@example.com"
)

type identityResponse struct {
	MerchantID string `json:"merchant_id"`
	Email      string `json:"email"`
}

func issuerWith(t *testing.T, secret string, ttl time.Duration) *commonjwt.Issuer {
	t.Helper()

	issuer, err := token.NewIssuer(&config.Config{JWT: config.JWTConfig{Secret: secret, TTL: ttl}})
	require.NoError(t, err)

	return issuer
}

func signedToken(t *testing.T, issuer *commonjwt.Issuer, claims commonjwt.Claims) string {
	t.Helper()

	signed, err := issuer.Issue(context.Background(), claims)
	require.NoError(t, err)

	return signed
}

func newGuardedApp(issuer *commonjwt.Issuer) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: response.Error})

	guarded := app.Group("/guarded", middleware.NewAuthMerchantMiddleware(issuer))
	guarded.Get("", func(c fiber.Ctx) error {
		return c.JSON(identityResponse{
			MerchantID: middleware.MerchantIDFromContext(c).String(),
			Email:      middleware.EmailFromContext(c),
		})
	})

	return app
}

func requestGuarded(t *testing.T, app *fiber.App, authorization string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/guarded", nil)
	if authorization != "" {
		req.Header.Set(fiber.HeaderAuthorization, authorization)
	}

	res, err := app.Test(req)
	require.NoError(t, err)

	return res
}

func TestNewAuthMerchantMiddlewareAcceptsAMerchantToken(t *testing.T) {
	issuer := issuerWith(t, testSecret, time.Hour)
	merchantID := uuid.New()

	raw := signedToken(t, issuer, commonjwt.Claims{
		UserID: merchantID,
		Email:  merchantEmail,
		Type:   domain.TokenTypeMerchant,
	})

	res := requestGuarded(t, newGuardedApp(issuer), "Bearer "+raw)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var identity identityResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&identity))

	assert.Equal(t, merchantID.String(), identity.MerchantID)
	assert.Equal(t, merchantEmail, identity.Email)
}

func TestNewAuthMerchantMiddlewareRejectsInvalidCredentials(t *testing.T) {
	issuer := issuerWith(t, testSecret, time.Hour)

	adminToken := signedToken(t, issuer, commonjwt.Claims{
		UserID: uuid.New(),
		Email:  "admin@example.com",
		Type:   domain.TokenTypeAdmin,
		Roles:  []string{"super_admin"},
	})

	foreignToken := signedToken(t, issuerWith(t, "another-service-secret", time.Hour), commonjwt.Claims{
		UserID: uuid.New(),
		Email:  merchantEmail,
		Type:   domain.TokenTypeMerchant,
	})

	expiredToken := signedToken(t, issuerWith(t, testSecret, -time.Hour), commonjwt.Claims{
		UserID: uuid.New(),
		Email:  merchantEmail,
		Type:   domain.TokenTypeMerchant,
	})

	tests := map[string]string{
		"no authorization header":      "",
		"empty bearer token":           "Bearer ",
		"missing bearer prefix":        signedToken(t, issuer, commonjwt.Claims{UserID: uuid.New(), Email: merchantEmail, Type: domain.TokenTypeMerchant}),
		"lowercase bearer prefix":      "bearer " + adminToken,
		"admin token":                  "Bearer " + adminToken,
		"token signed by another key":  "Bearer " + foreignToken,
		"expired token":                "Bearer " + expiredToken,
		"structurally malformed token": "Bearer not.a.jwt",
	}

	app := newGuardedApp(issuer)

	for name, authorization := range tests {
		t.Run(name, func(t *testing.T) {
			res := requestGuarded(t, app, authorization)

			assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		})
	}
}
