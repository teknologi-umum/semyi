package main

import "time"

// EnsureUTC ensures that a time.Time is in UTC timezone.
// If the time is not in UTC, it converts it to UTC.
// This is useful for database operations where we want to store all times in UTC.
func EnsureUTC(t time.Time) time.Time {
	if t.Location() != time.UTC {
		return t.UTC()
	}
	return t
}
