package json

import "time"

// JSONUtcTimestamp quickly creates a string RFC3339 format in UTC
func UtcTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// JSONUtcTimestampFromTime quickly creates a string RFC3339 format in UTC
func UtcTimestampFromTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
