package date_utils

import (
	"database/sql"
	"testing"
	"time"
)

func TestGetMidnightOfSeventhDay(t *testing.T) {
	now := time.Now().UTC()
	got := GetMidnightOfSeventhDay()

	want := time.Date(now.AddDate(0, 0, 7).Year(), now.AddDate(0, 0, 7).Month(), now.AddDate(0, 0, 7).Day(), 0, 0, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("GetMidnightOfSeventhDay() = %v, want %v", got, want)
	}

	if got.Location() != time.UTC {
		t.Fatalf("GetMidnightOfSeventhDay() location = %v, want UTC", got.Location())
	}
}

func TestFormatNullTime(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullTime
		expected string
	}{
		{name: "valid time", input: sql.NullTime{Time: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), Valid: true}, expected: "2024-01-02T03:04:05Z"},
		{name: "invalid time", input: sql.NullTime{}, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatNullTime(tt.input); got != tt.expected {
				t.Fatalf("FormatNullTime(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	input := time.Date(2024, 1, 2, 3, 4, 5, 123456789, time.FixedZone("UTC+2", 2*60*60))
	got := FormatTime(input)

	want := input.UTC().Format(time.RFC3339Nano)
	if got != want {
		t.Fatalf("FormatTime(%v) = %q, want %q", input, got, want)
	}
}
