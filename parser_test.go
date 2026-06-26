package years_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
	"github.com/expectto/be/be_time"
)

func TestParser_JustParseUnixTimestamp(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	var timestamp int64 = 1709682885

	parsedTime, err := years.DefaultParser().JustParse(strconv.Itoa(int(timestamp)))
	be.Require(t, err).To(be.Succeed())

	be.Expect(t, parsedTime).To(be_time.Unix(timestamp))
}

func TestParser_JustParseDateOnly(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	timeStr := "2024-03-06"
	expectedTime, _ := time.Parse(time.DateOnly, timeStr)

	years.SetParserDefaults(years.WithLayouts(time.DateOnly))

	parsedTime, err := years.DefaultParser().JustParse(timeStr)
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, parsedTime).To(be.Eq(expectedTime))
}

func TestParser_JustParseAliases(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	mockClock := &StaticClock{
		now: time.Date(2024, time.March, 01, 14, 30, 59, 0, time.UTC),
	}
	parser := years.NewParser(
		years.WithCustomClock(mockClock),
		years.AcceptAliases(),
		years.AcceptUnixSeconds(),
	)

	today, err := parser.JustParse("today")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, today.String()).To(be.Eq(`2024-03-01 00:00:00 +0000 UTC`))

	yesterday, err := parser.JustParse("yesterday")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, yesterday.String()).To(be.Eq(`2024-02-29 00:00:00 +0000 UTC`))

	tomorrow, err := parser.JustParse("tomorrow")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, tomorrow.String()).To(be.Eq(`2024-03-02 00:00:00 +0000 UTC`))
}
