package utils

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func FormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).Add(24*time.Hour - time.Nanosecond)
}
