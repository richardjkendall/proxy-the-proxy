package main

import (
	"testing"
	"time"
)

func TestUpdateMonthOfDate(t *testing.T) {
	year := int(time.Now().Year())
	date1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	date1 = UpdateMonthOfDate(date1, 3)
	if int(date1.Month()) != 3 {
		t.Fatalf("Month of date is not 3, value is = %v", date1)
	}
}

func TestUpdateDayOfDate(t *testing.T) {
	year := int(time.Now().Year())
	date1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	date1 = UpdateDayOfDate(date1, 15)
	if int(date1.Day()) != 15 {
		t.Fatalf("Month of date is not 15, value is = %v", date1)
	}
}

func TestUpdateYearOfDate(t *testing.T) {
	year := int(time.Now().Year())
	date1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	date1 = UpdateYearOfDate(date1, year+1)
	if int(date1.Year()) != year+1 {
		t.Fatalf("Month of date is not %v, value is = %v", year+1, date1)
	}
}
