package main

import (
	"time"
)

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

func UpdateSecondsOfTime(t time.Time, seconds int) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		seconds,
		t.Nanosecond(),
		t.Location(),
	)
}

func UpdateMinutesOfTime(t time.Time, minutes int) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		minutes,
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
}

func UpdateHoursOfTime(t time.Time, hours int) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		hours,
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
}

func DateLTE(d1 time.Time, d2 time.Time) bool {
	//log.Printf("DateLTE: d1 = %v, d2 = %v", d1, d2)
	return d1.Before(d2) || d1.Equal(d2)
}

func DateGTE(d1 time.Time, d2 time.Time) bool {
	return d1.After(d2) || d1.Equal(d2)
}
