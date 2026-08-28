package promptpay_test

import (
	"strings"
	"testing"

	"github.com/mudflap-autobotz/payment-transaction-service/pkg/utils/promptpay"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumMatchesCCITTFalseCheckValue(t *testing.T) {
	assert.Equal(t, "29B1", promptpay.CRC16ForTesting("123456789"))
}

func TestGenerateProducesScannablePayload(t *testing.T) {
	tests := []struct {
		name           string
		target         promptpay.Target
		amount         string
		reference      string
		expectedAmount string
		expectedTag29  string
	}{
		{
			name: "mobile target normalizes to 0066 prefix",
			target: promptpay.Target{
				Type:  promptpay.TargetTypeMobile,
				Value: "0812345678",
				Name:  "Payment Gateway",
				City:  "Bangkok",
			},
			amount:         "500",
			expectedAmount: "5406500.00",
			expectedTag29:  "29370016A000000677010111011300668123456780",
		},
		{
			name: "national id target",
			target: promptpay.Target{
				Type:  promptpay.TargetTypeNationalID,
				Value: "1234567890123",
				Name:  "Payment Gateway",
				City:  "Bangkok",
			},
			amount:         "1500.5",
			reference:      "ORDER-1",
			expectedAmount: "54071500.50",
			expectedTag29:  "29370016A00000067701011102131234567890123",
		},
		{
			name: "ewallet target",
			target: promptpay.Target{
				Type:  promptpay.TargetTypeEWallet,
				Value: "123456789012345",
				Name:  "Payment Gateway",
				City:  "Bangkok",
			},
			amount:         "100.00",
			expectedAmount: "5406100.00",
			expectedTag29:  "29390016A000000677010111031512345678901234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := promptpay.Generate(tt.target, tt.amount, tt.reference)
			require.NoError(t, err)

			assert.True(t, strings.HasPrefix(payload, "000201010212"), payload)
			assert.Contains(t, payload, tt.expectedAmount)
			assert.Contains(t, payload, "5303764")
			assert.Contains(t, payload, "5802TH")
			assert.Contains(t, payload, "5915PAYMENT GATEWAY")
			assert.Contains(t, payload, "6007BANGKOK")

			checksumIndex := len(payload) - 8
			require.Positive(t, checksumIndex)
			assert.Equal(t, "6304", payload[checksumIndex:checksumIndex+4])
			assert.Equal(t, promptpay.CRC16ForTesting(payload[:len(payload)-4]), payload[len(payload)-4:])
		})
	}
}

func TestGenerateEmbedsReferenceInAdditionalData(t *testing.T) {
	target := promptpay.Target{Type: promptpay.TargetTypeMobile, Value: "0812345678", Name: "Shop", City: "Bangkok"}

	t.Run("keeps reference case intact", func(t *testing.T) {
		payload, err := promptpay.Generate(target, "100.00", "Order-Abc")
		require.NoError(t, err)

		assert.Contains(t, payload, "62130109Order-Abc")
	})

	t.Run("omits additional data when reference is blank", func(t *testing.T) {
		payload, err := promptpay.Generate(target, "100.00", "   ")
		require.NoError(t, err)

		assert.NotContains(t, payload, "6213")
	})
}

func TestGenerateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		target      promptpay.Target
		amount      string
		expectedErr error
	}{
		{
			name:        "unsupported target type",
			target:      promptpay.Target{Type: "IBAN", Value: "0812345678"},
			amount:      "100.00",
			expectedErr: promptpay.ErrUnsupportedTargetType,
		},
		{
			name:        "mobile with wrong digit count",
			target:      promptpay.Target{Type: promptpay.TargetTypeMobile, Value: "08123456"},
			amount:      "100.00",
			expectedErr: promptpay.ErrInvalidTarget,
		},
		{
			name:        "national id with wrong digit count",
			target:      promptpay.Target{Type: promptpay.TargetTypeNationalID, Value: "12345"},
			amount:      "100.00",
			expectedErr: promptpay.ErrInvalidTarget,
		},
		{
			name:        "ewallet with wrong digit count",
			target:      promptpay.Target{Type: promptpay.TargetTypeEWallet, Value: "12345"},
			amount:      "100.00",
			expectedErr: promptpay.ErrInvalidTarget,
		},
		{
			name:        "zero amount",
			target:      promptpay.Target{Type: promptpay.TargetTypeMobile, Value: "0812345678"},
			amount:      "0",
			expectedErr: promptpay.ErrInvalidAmount,
		},
		{
			name:        "negative amount",
			target:      promptpay.Target{Type: promptpay.TargetTypeMobile, Value: "0812345678"},
			amount:      "-10.00",
			expectedErr: promptpay.ErrInvalidAmount,
		},
		{
			name:        "non numeric amount",
			target:      promptpay.Target{Type: promptpay.TargetTypeMobile, Value: "0812345678"},
			amount:      "abc",
			expectedErr: promptpay.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := promptpay.Generate(tt.target, tt.amount, "")

			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Empty(t, payload)
		})
	}
}

func TestGenerateAcceptsAlreadyPrefixedMobile(t *testing.T) {
	target := promptpay.Target{Type: promptpay.TargetTypeMobile, Value: "0066812345678", Name: "Shop", City: "Bangkok"}

	payload, err := promptpay.Generate(target, "100.00", "")
	require.NoError(t, err)

	assert.Contains(t, payload, "0113006681234567")
}
