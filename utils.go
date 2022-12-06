package main

import "time"

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

func UpdateYearOfDate(d time.Time, year int) time.Time {
	return time.Date(
		year,
		d.Month(),
		d.Day(),
		d.Hour(),
		d.Minute(),
		d.Second(),
		d.Nanosecond(),
		d.Location(),
	)
}

func UpdateMonthOfDate(d time.Time, month int) time.Time {
	return time.Date(
		d.Year(),
		time.Month(month),
		d.Day(),
		d.Hour(),
		d.Minute(),
		d.Second(),
		d.Nanosecond(),
		d.Location(),
	)
}

func UpdateDayOfDate(d time.Time, day int) time.Time {
	return time.Date(
		d.Year(),
		d.Month(),
		day,
		d.Hour(),
		d.Minute(),
		d.Second(),
		d.Nanosecond(),
		d.Location(),
	)
}

func DateLTE(d1 time.Time, d2 time.Time) bool {
	return d1.Before(d2) || d1.Equal(d2)
}

func DateGTE(d1 time.Time, d2 time.Time) bool {
	return d1.After(d2) || d1.Equal(d2)
}
