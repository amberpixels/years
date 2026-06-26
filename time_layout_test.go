package years_test

import (
	"fmt"
	"testing"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
)

func TestParseLayout_ValidGoLayouts(t *testing.T) {
	tests := []struct {
		name          string
		layout        string
		expectedUnit  years.DateUnit
		expectedUnits []years.DateUnit
	}{
		{"full date", "2006-01-02", years.Day, []years.DateUnit{years.Day, years.Month, years.Year}},
		{"year-month", "2006-01", years.Month, []years.DateUnit{years.Month, years.Year}},
		{"day-month-year", "02-01-2006", years.Day, []years.DateUnit{years.Day, years.Month, years.Year}},
		{"year only", "2006", years.Year, []years.DateUnit{years.Year}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			details := years.ParseLayout(tc.layout)
			be.Require(t, details).NotTo(be.Nil(), fmt.Sprintf("ParseLayout should not return nil for %s", tc.layout))
			be.Expect(t, details.Format).To(be.Eq(years.LayoutFormatGo))
			be.Expect(t, details.MinimalUnit).To(be.Eq(tc.expectedUnit))
			be.Expect(t, details.Units).To(be.Eq(tc.expectedUnits))
		})
	}
}

func TestParseLayout_UnixTimestampLayouts(t *testing.T) {
	tests := []struct {
		name         string
		layout       string
		expectedUnit years.DateUnit
	}{
		{"seconds", years.LayoutTimestampSeconds, years.UnixSecond},
		{"milliseconds", years.LayoutTimestampMilliseconds, years.UnixMillisecond},
		{"microseconds", years.LayoutTimestampMicroseconds, years.UnixMicrosecond},
		{"nanoseconds", years.LayoutTimestampNanoseconds, years.UnixNanosecond},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			details := years.ParseLayout(tc.layout)
			be.Require(t, details).NotTo(be.Nil(), fmt.Sprintf("ParseLayout should not return nil for %s", tc.layout))
			be.Expect(t, details.Format).To(be.Eq(years.LayoutFormatUnixTimestamp))
			be.Expect(t, details.MinimalUnit).To(be.Eq(tc.expectedUnit))
			be.Expect(t, details.Units).To(be.Eq([]years.DateUnit{tc.expectedUnit}))
		})
	}
}

func TestParseLayout_UnknownLayout(t *testing.T) {
	be.Expect(t, years.ParseLayout("foo-bar")).To(be.Nil())
}
