package money_test

import (
	"testing"

	"github.com/mudflap-autobotz/payment-transaction-service/pkg/utils/money"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	t.Run("pads to two decimal places", func(t *testing.T) {
		tests := map[string]string{
			"100":           "100.00",
			"100.5":         "100.50",
			"100.50":        "100.50",
			"  250.25  ":    "250.25",
			"0.01":          "0.01",
			"1234567890123": "1234567890123.00",
		}

		for raw, expected := range tests {
			normalized, err := money.Normalize(raw)

			require.NoError(t, err, raw)
			assert.Equal(t, expected, normalized, raw)
		}
	})

	t.Run("rejects invalid amounts", func(t *testing.T) {
		invalid := []string{
			"",
			"   ",
			"abc",
			"0",
			"0.00",
			"-1.00",
			"+1.00",
			"1e5",
			"1E5",
			"3/4",
			"100.001",
			"12345678901234",
		}

		for _, raw := range invalid {
			normalized, err := money.Normalize(raw)

			assert.ErrorIs(t, err, money.ErrInvalidAmount, raw)
			assert.Empty(t, normalized, raw)
		}
	})
}

func TestCompare(t *testing.T) {
	t.Run("compares exact decimals", func(t *testing.T) {
		result, err := money.Compare("100.10", "100.20")
		require.NoError(t, err)
		assert.Negative(t, result)

		result, err = money.Compare("100.20", "100.10")
		require.NoError(t, err)
		assert.Positive(t, result)

		result, err = money.Compare("100.00", "100")
		require.NoError(t, err)
		assert.Zero(t, result)
	})

	t.Run("does not lose precision the way float64 would", func(t *testing.T) {
		result, err := money.Compare("0.1", "0.10")
		require.NoError(t, err)
		assert.Zero(t, result)

		result, err = money.Compare("1234567890123.45", "1234567890123.46")
		require.NoError(t, err)
		assert.Negative(t, result)
	})

	t.Run("rejects unparsable operands", func(t *testing.T) {
		_, err := money.Compare("abc", "100.00")
		assert.ErrorIs(t, err, money.ErrInvalidAmount)

		_, err = money.Compare("100.00", "abc")
		assert.ErrorIs(t, err, money.ErrInvalidAmount)
	})
}

func TestInRange(t *testing.T) {
	const minimum = "100.00"
	const maximum = "1000000.00"

	t.Run("boundaries are inclusive", func(t *testing.T) {
		for _, amount := range []string{minimum, maximum, "500.00"} {
			inRange, err := money.InRange(amount, minimum, maximum)

			require.NoError(t, err, amount)
			assert.True(t, inRange, amount)
		}
	})

	t.Run("outside the range", func(t *testing.T) {
		for _, amount := range []string{"99.99", "1000000.01"} {
			inRange, err := money.InRange(amount, minimum, maximum)

			require.NoError(t, err, amount)
			assert.False(t, inRange, amount)
		}
	})

	t.Run("rejects unparsable bounds", func(t *testing.T) {
		_, err := money.InRange("500.00", "abc", maximum)
		assert.ErrorIs(t, err, money.ErrInvalidAmount)

		_, err = money.InRange("500.00", minimum, "abc")
		assert.ErrorIs(t, err, money.ErrInvalidAmount)
	})
}
