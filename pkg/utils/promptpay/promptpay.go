package promptpay

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	TargetTypeMobile     = "MOBILE"
	TargetTypeNationalID = "NATID"
	TargetTypeEWallet    = "EWALLET"
)

const (
	tagPayloadFormat     = "00"
	tagPointOfInitiation = "01"
	tagMerchantAccount   = "29"
	tagCurrency          = "53"
	tagAmount            = "54"
	tagCountry           = "58"
	tagMerchantName      = "59"
	tagMerchantCity      = "60"
	tagAdditionalData    = "62"
	tagChecksum          = "63"

	subTagAcquirerID  = "00"
	subTagMobile      = "01"
	subTagNationalID  = "02"
	subTagEWallet     = "03"
	subTagBillPayment = "01"

	payloadFormatVersion = "01"
	dynamicQR            = "12"
	promptPayAcquirerID  = "A000000677010111"
	currencyTHB          = "764"
	countryTH            = "TH"

	mobileCountryCode = "0066"

	maxMerchantNameLength = 25
	maxMerchantCityLength = 15
	maxReferenceLength    = 25

	nationalIDLength     = 13
	eWalletIDLength      = 15
	localMobileLength    = 10
	prefixedMobileLength = 13
)

var (
	ErrUnsupportedTargetType = errors.New("unsupported promptpay target type")
	ErrInvalidTarget         = errors.New("invalid promptpay target value")
	ErrInvalidAmount         = errors.New("invalid promptpay amount")
)

type Target struct {
	Type  string
	Value string
	Name  string
	City  string
}

func Generate(target Target, amount, reference string) (string, error) {
	accountInfo, err := merchantAccountInfo(target)
	if err != nil {
		return "", err
	}

	formattedAmount, err := formatAmount(amount)
	if err != nil {
		return "", err
	}

	var payload strings.Builder

	payload.WriteString(tlv(tagPayloadFormat, payloadFormatVersion))
	payload.WriteString(tlv(tagPointOfInitiation, dynamicQR))
	payload.WriteString(tlv(tagMerchantAccount, accountInfo))
	payload.WriteString(tlv(tagCurrency, currencyTHB))
	payload.WriteString(tlv(tagAmount, formattedAmount))
	payload.WriteString(tlv(tagCountry, countryTH))
	payload.WriteString(tlv(tagMerchantName, sanitize(target.Name, maxMerchantNameLength)))
	payload.WriteString(tlv(tagMerchantCity, sanitize(target.City, maxMerchantCityLength)))

	if normalizedReference := truncate(strings.TrimSpace(reference), maxReferenceLength); normalizedReference != "" {
		payload.WriteString(tlv(tagAdditionalData, tlv(subTagBillPayment, normalizedReference)))
	}

	payload.WriteString(tagChecksum)
	payload.WriteString("04")

	return payload.String() + crc16(payload.String()), nil
}

func merchantAccountInfo(target Target) (string, error) {
	digits := onlyDigits(target.Value)

	switch strings.ToUpper(strings.TrimSpace(target.Type)) {
	case TargetTypeMobile:
		mobile, err := normalizeMobile(digits)
		if err != nil {
			return "", err
		}

		return tlv(subTagAcquirerID, promptPayAcquirerID) + tlv(subTagMobile, mobile), nil

	case TargetTypeNationalID:
		if len(digits) != nationalIDLength {
			return "", ErrInvalidTarget
		}

		return tlv(subTagAcquirerID, promptPayAcquirerID) + tlv(subTagNationalID, digits), nil

	case TargetTypeEWallet:
		if len(digits) != eWalletIDLength {
			return "", ErrInvalidTarget
		}

		return tlv(subTagAcquirerID, promptPayAcquirerID) + tlv(subTagEWallet, digits), nil

	default:
		return "", ErrUnsupportedTargetType
	}
}

func normalizeMobile(digits string) (string, error) {
	switch {
	case len(digits) == localMobileLength && strings.HasPrefix(digits, "0"):
		return mobileCountryCode + digits[1:], nil

	case len(digits) == prefixedMobileLength && strings.HasPrefix(digits, mobileCountryCode):
		return digits, nil

	default:
		return "", ErrInvalidTarget
	}
}

func formatAmount(amount string) (string, error) {
	value, ok := new(big.Rat).SetString(strings.TrimSpace(amount))
	if !ok || value.Sign() <= 0 {
		return "", ErrInvalidAmount
	}

	return value.FloatString(2), nil
}

func tlv(id, value string) string {
	return fmt.Sprintf("%s%02d%s", id, len(value), value)
}

func sanitize(value string, maxLength int) string {
	return truncate(strings.ToUpper(strings.TrimSpace(value)), maxLength)
}

func truncate(value string, maxLength int) string {
	if len(value) > maxLength {
		return value[:maxLength]
	}

	return value
}

func onlyDigits(value string) string {
	var digits strings.Builder

	for _, character := range value {
		if character >= '0' && character <= '9' {
			digits.WriteRune(character)
		}
	}

	return digits.String()
}
