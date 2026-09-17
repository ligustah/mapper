package example

import (
	"strconv"
	"strings"
	"time"
)

// ConvertTimestamp is a custom converter function for the using: directive
// It converts int64 (unix timestamp) to string
func ConvertTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format(time.RFC3339)
}

// ParseTimestamp is a custom converter function for the using: directive
// It converts string (RFC3339) to int64 (unix timestamp)
func ParseTimestamp(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// IntToString converts int to string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// FormatTime is a custom converter for time.Time -> string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FromNanos is a custom converter for int64 (unix nanoseconds) -> time.Time
func FromNanos(ns int64) time.Time {
	return time.Unix(0, ns).UTC()
}

// NanosOrNil is a custom converter for int64 (unix nanoseconds) -> *time.Time,
// nil for zero
func NanosOrNil(ns int64) *time.Time {
	if ns == 0 {
		return nil
	}

	t := time.Unix(0, ns).UTC()
	return &t
}

// Upper is a custom converter for string -> string
func Upper(s string) string {
	return strings.ToUpper(s)
}

// ParseTime is a custom converter for string -> time.Time
func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
