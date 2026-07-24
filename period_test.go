package years_test

import (
	"testing"
	"time"

	"github.com/expectto/be"

	"github.com/amberpixels/years"
)

// periodTestParser builds a parser with a deterministic clock and the same
// capabilities the default parser has.
func periodTestParser(now time.Time) *years.Parser {
	return years.NewParser(
		years.WithCustomClock(&StaticClock{now: now}),
		years.AcceptAliases(),
		years.AcceptRelativeOffsets(),
		years.AcceptUnixSeconds(),
		years.WithLayouts(years.DefaultLayouts...),
	)
}

func TestParsePeriod_Aliases(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	// Friday, mid-day, so weekday-relative aliases are unambiguous.
	now := time.Date(2024, time.March, 1, 14, 30, 59, 0, time.UTC)
	parser := periodTestParser(now)

	day := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	cases := map[string]years.Period{
		"today":     {Start: day(2024, time.March, 1), End: day(2024, time.March, 2)},
		"yesterday": {Start: day(2024, time.February, 29), End: day(2024, time.March, 1)},
		"tomorrow":  {Start: day(2024, time.March, 2), End: day(2024, time.March, 3)},
		// Weeks start on Sunday: Feb 25 begins the current week.
		"this-week":  {Start: day(2024, time.February, 25), End: day(2024, time.March, 3)},
		"last-week":  {Start: day(2024, time.February, 18), End: day(2024, time.February, 25)},
		"this-month": {Start: day(2024, time.March, 1), End: day(2024, time.April, 1)},
		"last-month": {Start: day(2024, time.February, 1), End: day(2024, time.March, 1)},
		"this-year":  {Start: day(2024, time.January, 1), End: day(2025, time.January, 1)},
		"last-year":  {Start: day(2023, time.January, 1), End: day(2024, time.January, 1)},
		// Weekends are two days from Saturday 00:00.
		"next-weekend": {Start: day(2024, time.March, 2), End: day(2024, time.March, 4)},
		"last-weekend": {Start: day(2024, time.February, 24), End: day(2024, time.February, 26)},
	}
	for in, want := range cases {
		got, err := parser.JustParsePeriod(in)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got.Start.String()).To(be.Eq(want.Start.String()), "start of %q", in)
		be.Expect(t, got.End.String()).To(be.Eq(want.End.String()), "end of %q", in)
	}
}

func TestParsePeriod_RelativeOffsets(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	now := time.Date(2024, time.March, 15, 10, 0, 0, 0, time.UTC)
	parser := periodTestParser(now)

	cases := map[string]time.Time{
		"12h": now.Add(-12 * time.Hour),
		"7d":  now.AddDate(0, 0, -7),
		"2w":  now.AddDate(0, 0, -14),
		"3mo": now.AddDate(0, -3, 0),
		"1y":  now.AddDate(-1, 0, 0),
	}
	for in, wantStart := range cases {
		got, err := parser.JustParsePeriod(in)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got.Start.String()).To(be.Eq(wantStart.String()), "start of %q", in)
		be.Expect(t, got.End.String()).To(be.Eq(now.String()), "end of %q", in)
	}

	// The instant Parse returns the offset period's start.
	instant, err := parser.JustParse("7d")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, instant.String()).To(be.Eq(now.AddDate(0, 0, -7).String()))
}

func TestParsePeriod_Layouts(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	parser := periodTestParser(time.Date(2024, time.March, 15, 10, 0, 0, 0, time.UTC))

	// A date-only value denotes the whole day.
	day, err := parser.JustParsePeriod("2024-03-06")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, day.Start.String()).To(be.Eq("2024-03-06 00:00:00 +0000 UTC"))
	be.Expect(t, day.End.String()).To(be.Eq("2024-03-07 00:00:00 +0000 UTC"))
	be.Expect(t, day.IsInstant()).To(be.Eq(false))

	// A timestamp denotes an instant (zero-length period).
	instant, err := parser.JustParsePeriod("2024-03-06T14:30:00Z")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, instant.IsInstant()).To(be.Eq(true))
	be.Expect(t, instant.Start.String()).To(be.Eq("2024-03-06 14:30:00 +0000 UTC"))

	// Unix epochs are instants too.
	epoch, err := parser.JustParsePeriod("1709682885")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, epoch.IsInstant()).To(be.Eq(true))

	// Garbage still fails.
	_, err = parser.JustParsePeriod("not-a-time")
	be.Expect(t, err).To(be.HaveOccurred())
}

func TestPeriod_Contains(t *testing.T) {
	start := time.Date(2024, time.March, 6, 0, 0, 0, 0, time.UTC)
	p := years.Period{Start: start, End: start.AddDate(0, 0, 1)}

	be.Expect(t, p.Contains(start)).To(be.Eq(true))
	be.Expect(t, p.Contains(start.Add(23*time.Hour))).To(be.Eq(true))
	be.Expect(t, p.Contains(p.End)).To(be.Eq(false)) // half-open
	be.Expect(t, p.Duration()).To(be.Eq(24 * time.Hour))

	instant := years.InstantPeriod(start)
	be.Expect(t, instant.Contains(start)).To(be.Eq(true))
	be.Expect(t, instant.Contains(start.Add(time.Second))).To(be.Eq(false))
}

func TestParsePeriod_PackageLevel(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	// The default parser accepts periods with no per-call setup.
	p, err := years.JustParsePeriod("2024-03-06")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, p.Duration()).To(be.Eq(24 * time.Hour))
}
