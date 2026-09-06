package dto

import "time"

type DateRangeQuery struct {
	DateFrom string `query:"date_from" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	DateTo   string `query:"date_to" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

func OptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
