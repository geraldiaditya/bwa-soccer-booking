package utils

import (
	"testing"
	"time"
)

func TestIndonesianMonthName(t *testing.T) {
	tests := []struct {
		month time.Month
		want  string
	}{
		{time.January, "Januari"},
		{time.February, "Februari"},
		{time.March, "Maret"},
		{time.April, "April"},
		{time.May, "Mei"},
		{time.June, "Juni"},
		{time.July, "Juli"},
		{time.August, "Agustus"},
		{time.September, "September"},
		{time.October, "Oktober"},
		{time.November, "November"},
		{time.December, "Desember"},
	}

	for _, tt := range tests {
		if got := IndonesianMonthName(tt.month); got != tt.want {
			t.Fatalf("IndonesianMonthName(%s) = %q, want %q", tt.month, got, tt.want)
		}
	}
}

func TestFormatIndonesianDate(t *testing.T) {
	date := time.Date(2026, time.May, 9, 15, 4, 0, 0, time.UTC)

	got := FormatIndonesianDate(date)
	if got != "09 Mei 2026" {
		t.Fatalf("FormatIndonesianDate() = %q, want %q", got, "09 Mei 2026")
	}
}

func TestConvertMonthName(t *testing.T) {
	tests := []struct {
		name string
		date string
		want string
	}{
		{name: "may", date: "2026-05-09", want: "09 Mei"},
		{name: "october", date: "2026-10-09", want: "09 Okt"},
		{name: "invalid", date: "bad-date", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertMonthName(tt.date); got != tt.want {
				t.Fatalf("ConvertMonthName(%q) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}
