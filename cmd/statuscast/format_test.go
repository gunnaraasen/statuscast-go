package main

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{
			"2024-03-15T10:30:00Z",
			time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			"2024-03-15T10:30",
			time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			"2024-03-15",
			time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		got, err := parseTime(tt.input)
		if err != nil {
			t.Errorf("parseTime(%q) error: %v", tt.input, err)
			continue
		}
		if !got.Equal(tt.want) {
			t.Errorf("parseTime(%q) = %v; want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseTime_Invalid(t *testing.T) {
	_, err := parseTime("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "-"},
		{45 * time.Second, "45s"},
		{2*time.Minute + 30*time.Second, "2m30s"},
		{3*time.Hour + 5*time.Minute + 10*time.Second, "3h5m10s"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q; want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatTime_Zero(t *testing.T) {
	if got := formatTime(time.Time{}); got != "-" {
		t.Errorf("formatTime(zero) = %q; want %q", got, "-")
	}
}

func TestFormatOptTime_Nil(t *testing.T) {
	if got := formatOptTime(nil); got != "-" {
		t.Errorf("formatOptTime(nil) = %q; want %q", got, "-")
	}
}

func TestFormatOptTime_NonNil(t *testing.T) {
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	want := "2024-01-02 03:04:05"
	if got := formatOptTime(&ts); got != want {
		t.Errorf("formatOptTime(&ts) = %q; want %q", got, want)
	}
}
