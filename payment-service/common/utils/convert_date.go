package utils

import (
	"fmt"
	"time"
)

func IndonesianMonthName(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}
	return months[month]
}

func FormatIndonesianDate(date time.Time) string {
	return fmt.Sprintf("%02d %s %d", date.Day(), IndonesianMonthName(date.Month()), date.Year())
}

func ConvertMonthName(inputDate string) string {
	date, err := time.Parse(time.DateOnly, inputDate)
	if err != nil {
		return ""
	}
	months := map[time.Month]string{
		time.January:   "Jan",
		time.February:  "Feb",
		time.March:     "Mar",
		time.April:     "Apr",
		time.May:       "Mei",
		time.June:      "Jun",
		time.July:      "Jul",
		time.August:    "Agu",
		time.September: "Sep",
		time.October:   "Okt",
		time.November:  "Nov",
		time.December:  "Des",
	}
	return fmt.Sprintf("%02d %s", date.Day(), months[date.Month()])
}
