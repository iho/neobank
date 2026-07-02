package api

import "time"

func parseTimeOrNow(value string) time.Time {
	if value == "" {
		return time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Now().UTC()
}

func parseTimePtr(value string) *time.Time {
	if value == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t
	}
	return nil
}