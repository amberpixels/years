package years_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/expectto/be"
	"github.com/expectto/be/be_time"

	"github.com/amberpixels/years"
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

func TestParser_DefaultLayouts(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	// The default parser handles common ISO-8601 forms with no per-call setup.
	// The last two cases prove RFC3339Nano covers timestamps both without and
	// with a fractional second, so a separate RFC3339 layout is unnecessary.
	cases := map[string]time.Time{
		"2024-03-06":             time.Date(2024, time.March, 6, 0, 0, 0, 0, time.UTC),
		"2024-03-06 14:30:00":    time.Date(2024, time.March, 6, 14, 30, 0, 0, time.UTC),
		"2024-03-06 14:30":       time.Date(2024, time.March, 6, 14, 30, 0, 0, time.UTC),
		"2024-03-06T14:30:00Z":   time.Date(2024, time.March, 6, 14, 30, 0, 0, time.UTC),
		"2024-03-06T14:30:00.5Z": time.Date(2024, time.March, 6, 14, 30, 0, 500000000, time.UTC),
	}
	for in, want := range cases {
		got, err := years.JustParse(in)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got.UTC()).To(be.Eq(want))
	}
}

func TestParser_JustParseAliases(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	mockClock := &StaticClock{
		now: time.Date(2024, time.March, 0o1, 14, 30, 59, 0, time.UTC),
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

func TestParser_WithCustomAliases(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	mockClock := &StaticClock{
		now: time.Date(2024, time.March, 0o1, 14, 30, 59, 0, time.UTC),
	}
	customAliases := map[string]func(time.Time) time.Time{
		"day-after-tomorrow": func(base time.Time) time.Time {
			base = base.AddDate(0, 0, 2)
			return years.Mutate(&base).TruncateToDay().Time()
		},
	}

	parser := years.NewParser(
		years.WithCustomClock(mockClock),
		years.AcceptAliases(),
		years.WithCustomAliases(customAliases),
	)

	// The custom alias resolves
	dayAfterTomorrow, err := parser.JustParse("day-after-tomorrow")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, dayAfterTomorrow.String()).To(be.Eq(`2024-03-03 00:00:00 +0000 UTC`))

	// Core aliases stay intact on the same parser
	today, err := parser.JustParse("today")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, today.String()).To(be.Eq(`2024-03-01 00:00:00 +0000 UTC`))

	// The caller's map was not polluted with core aliases
	be.Expect(t, customAliases).To(be.HaveLength(1))

	// The shared core alias table was not polluted either: a fresh parser
	// without the option must not know the custom alias
	freshParser := years.NewParser(years.WithCustomClock(mockClock), years.AcceptAliases())
	_, err = freshParser.JustParse("day-after-tomorrow")
	be.Expect(t, err).To(be.HaveOccurred())
}

func TestParser_WithCustomAliases_OverridesAndOrder(t *testing.T) {
	// Friday.
	mockClock := &StaticClock{
		now: time.Date(2024, time.March, 0o1, 14, 30, 59, 0, time.UTC),
	}
	fixed := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	// The week start given after the custom aliases still applies, and a
	// custom alias overrides the same-named core one.
	parser := years.NewParser(
		years.WithCustomClock(mockClock),
		years.AcceptAliases(),
		years.WithCustomAliases(map[string]func(time.Time) time.Time{
			"today": func(time.Time) time.Time { return fixed },
		}),
		years.WithWeekStart(time.Monday),
	)

	today, err := parser.JustParse("today")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, today.String()).To(be.Eq(fixed.String()))

	thisWeek, err := parser.JustParse("this-week")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, thisWeek.String()).To(be.Eq(`2024-02-26 00:00:00 +0000 UTC`))
}

func TestParser_Wildcards(t *testing.T) {
	jsDate := time.Date(2025, time.April, 24, 11, 52, 5, 0, time.FixedZone("", 3*60*60))
	april24 := time.Date(2025, time.April, 24, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		layout string
		value  string
		want   time.Time
	}{
		{
			"Mon Jan 02 2006 15:04:05 GMT-0700*",
			"Thu Apr 24 2025 11:52:05 GMT+0300 (Eastern European Summer Time)",
			jsDate,
		},
		{"2006-01-02*", "2025-04-24 (Thursday)", april24},
		{"2006-01-02*", "2025-04-24", april24},
		{"*2006-01-02", "created: 2025-04-24", april24},
		{"*2006-01-02*", "[2025-04-24]", april24},
		{"*2006-01-02*", "2025-04-24", april24},
		// The longest prefix keeps the whole day, not "2025-4-2".
		{"2006-1-2*", "2025-4-24 x", april24},
		// The shortest skip keeps the whole month, not "1/2/2025".
		{"*1/2/2006", "x 11/2/2025", time.Date(2025, time.November, 2, 0, 0, 0, 0, time.UTC)},
		{"*2006-1-2*", "[2025-4-24]", april24},
		{"*2006-01-02*", "дата: 2025-04-24 (четверг)", april24},
	}
	parser := years.NewParser()
	for _, tc := range cases {
		got, err := parser.Parse(tc.layout, tc.value)
		be.Require(t, err).To(be.Succeed(), "%q with %q", tc.value, tc.layout)
		be.Expect(t, got.Equal(tc.want)).To(be.Eq(true), "%q with %q: got %s", tc.value, tc.layout, got)
		_, gotOffset := got.Zone()
		_, wantOffset := tc.want.Zone()
		be.Expect(t, gotOffset).To(be.Eq(wantOffset), "%q with %q", tc.value, tc.layout)
	}
}

func TestParser_WildcardErrors(t *testing.T) {
	cases := []struct {
		layout string
		value  string
	}{
		{"2006*01-02", "2025*04-24"},
		{"**2006-01-02", "2025-04-24"},
		{"2006-01-02**", "2025-04-24"},
		{"*", "2025-04-24"},
		{"2006-01-02*", "no date here"},
		{"*2006-01-02*", "no date here"},
	}
	parser := years.NewParser()
	for _, tc := range cases {
		_, err := parser.Parse(tc.layout, tc.value)
		be.Expect(t, err).To(be.HaveOccurred(), "%q with %q", tc.value, tc.layout)
	}

	_, err := parser.Parse("2006*01-02", "2025*04-24")
	be.Expect(t, err.Error()).To(be.ContainSubstring("layout(2006*01-02)"))
}

func TestParser_WildcardAfterStrictLayouts(t *testing.T) {
	parser := years.NewParser(years.WithLayouts(time.RFC3339, "2006-01-02*"))

	got, err := parser.JustParse("2025-04-24T10:00:00Z")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, got).To(be.Eq(time.Date(2025, time.April, 24, 10, 0, 0, 0, time.UTC)))

	got, err = parser.JustParse("2025-04-24 (Thursday)")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, got).To(be.Eq(time.Date(2025, time.April, 24, 0, 0, 0, 0, time.UTC)))
}
