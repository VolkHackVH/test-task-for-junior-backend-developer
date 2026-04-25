package handlers

import (
	"fmt"
	"time"
)

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func parseDates(values []string) ([]time.Time, error) {
	dates := make([]time.Time, 0, len(values))

	for _, value := range values {
		parsed, err := parseDate(value)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: expected format \"YYYY-MM-DD\"", value)
		}

		dates = append(dates, parsed)
	}

	return dates, nil
}
