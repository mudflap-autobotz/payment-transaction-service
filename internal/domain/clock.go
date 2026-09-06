package domain

import "time"

func StartOfDay(now time.Time, location *time.Location) time.Time {
	local := now.In(location)

	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}
