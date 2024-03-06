package years_test

import (
	years "github.com/amberpixels/era"
	"testing"
	"time"
)

func TestParseTimeUnixTimestamp(t *testing.T) {
	timeStr := "1709682885"
	expectedTime := time.Unix(1709682885, 0)

	parsedTime, err := years.ParseTime(timeStr)
	if err != nil {
		t.Errorf("Error parsing time: %v", err)
	}

	if !parsedTime.Equal(expectedTime) {
		t.Errorf("Parsed time doesn't match expected time. Expected: %v, Got: %v", expectedTime, parsedTime)
	}
}

func TestParseTimeDateOnly(t *testing.T) {
	timeStr := "2024-03-06"
	expectedTime, _ := time.Parse(time.DateOnly, timeStr)

	years.SetDefaults(years.WithLayouts(time.DateOnly))

	parsedTime, err := years.ParseTime(timeStr)
	if err != nil {
		t.Errorf("Error parsing time: %v", err)
	}

	if !parsedTime.Equal(expectedTime) {
		t.Errorf("Parsed time doesn't match expected time. Expected: %v, Got: %v", expectedTime, parsedTime)
	}
}
