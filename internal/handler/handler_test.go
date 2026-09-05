package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonjwt "github.com/mudflap-autobotz/payment-common/jwt"
	"github.com/mudflap-autobotz/payment-common/jwt/jwttest"
	"github.com/mudflap-autobotz/payment-common/response"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/config"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/domain"
	"github.com/mudflap-autobotz/payment-transaction-service/internal/token"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var errDatabase = errors.New("database error")

const (
	merchantEmail = "shop@example.com"
	invalidUUID   = "not-a-uuid"
)

var handlerTestKeys = jwttest.MustGenerateKeys()

func newApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: response.Error})
}

func testSigner(t *testing.T) *commonjwt.Signer {
	t.Helper()

	signer, err := commonjwt.NewSigner(commonjwt.SignerConfig{
		PrivateKey: handlerTestKeys.PrivatePEM,
		TTL:        time.Hour,
	})
	require.NoError(t, err)

	return signer
}

func testVerifier(t *testing.T) *commonjwt.Verifier {
	t.Helper()

	verifier, err := token.NewVerifier(&config.Config{JWT: config.JWTConfig{PublicKey: handlerTestKeys.PublicPEM}})
	require.NoError(t, err)

	return verifier
}

func bearerFor(t *testing.T, subjectID uuid.UUID, email, tokenType string) string {
	t.Helper()

	signed, err := testSigner(t).Issue(context.Background(), commonjwt.Claims{
		UserID: subjectID,
		Email:  email,
		Type:   tokenType,
	})
	require.NoError(t, err)

	return "Bearer " + signed
}

func merchantBearer(t *testing.T, merchantID uuid.UUID) string {
	t.Helper()

	return bearerFor(t, merchantID, merchantEmail, domain.TokenTypeMerchant)
}

func newJSONRequest(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, target, reader)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return req
}

func doRequest(t *testing.T, app *fiber.App, method, target string, body any) *http.Response {
	t.Helper()

	res, err := app.Test(newJSONRequest(t, method, target, body))
	require.NoError(t, err)

	t.Cleanup(func() { _ = res.Body.Close() })

	return res
}

func doAuthorizedRequest(t *testing.T, app *fiber.App, method, target, authorization string, body any) *http.Response {
	t.Helper()

	req := newJSONRequest(t, method, target, body)
	req.Header.Set(fiber.HeaderAuthorization, authorization)

	res, err := app.Test(req)
	require.NoError(t, err)

	t.Cleanup(func() { _ = res.Body.Close() })

	return res
}

func decodeBody[T any](t *testing.T, res *http.Response) T {
	t.Helper()

	var body T
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	return body
}

type dataEnvelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type paginateEnvelope[T any] struct {
	Code       int                 `json:"code"`
	Message    string              `json:"message"`
	Data       T                   `json:"data"`
	Pagination response.Pagination `json:"pagination"`
}

type errorEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
