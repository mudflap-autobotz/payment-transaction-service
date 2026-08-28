package money

import (
	"errors"
	"math/big"
	"strings"
)

const (
	Scale       = 2
	maxIntegers = 13
)

var ErrInvalidAmount = errors.New("amount must be a positive decimal with at most 2 decimal places")

func Normalize(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidAmount
	}

	if strings.ContainsAny(trimmed, "eE/+ ") {
		return "", ErrInvalidAmount
	}

	value, ok := new(big.Rat).SetString(trimmed)
	if !ok || value.Sign() <= 0 {
		return "", ErrInvalidAmount
	}

	normalized := value.FloatString(Scale)

	if roundTrip, _ := new(big.Rat).SetString(normalized); roundTrip.Cmp(value) != 0 {
		return "", ErrInvalidAmount
	}

	if integers, _, _ := strings.Cut(normalized, "."); len(integers) > maxIntegers {
		return "", ErrInvalidAmount
	}

	return normalized, nil
}

func Compare(left, right string) (int, error) {
	leftValue, ok := new(big.Rat).SetString(strings.TrimSpace(left))
	if !ok {
		return 0, ErrInvalidAmount
	}

	rightValue, ok := new(big.Rat).SetString(strings.TrimSpace(right))
	if !ok {
		return 0, ErrInvalidAmount
	}

	return leftValue.Cmp(rightValue), nil
}

func InRange(amount, minimum, maximum string) (bool, error) {
	aboveMinimum, err := Compare(amount, minimum)
	if err != nil {
		return false, err
	}

	belowMaximum, err := Compare(amount, maximum)
	if err != nil {
		return false, err
	}

	return aboveMinimum >= 0 && belowMaximum <= 0, nil
}
