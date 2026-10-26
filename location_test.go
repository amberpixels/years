package years_test

import (
	"testing"
	"time"

	"github.com/expectto/be"

	"github.com/amberpixels/years"
)

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	be.Require(t, err).To(be.Succeed())
	return loc
}

func TestWithLocation_Layouts(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	chisinau := mustLoadLocation(t, "Europe/Chisinau")
	chicago := mustLoadLocation(t, "America/Chicago")
	now := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		loc   *time.Location
		value string
		want  time.Time
	}{
		{chisinau, "2026-10-20", time.Date(2026, time.October, 20, 0, 0, 0, 0, chisinau)},
		{chisinau, "2026-10-20 09:00", time.Date(2026, time.October, 20, 9, 0, 0, 0, chisinau)},
		{chisinau, "2026-10-20 09:00:00", time.Date(2026, time.October, 20, 9, 0, 0, 0, chisinau)},
		{chicago, "2026-10-20", time.Date(2026, time.October, 20, 0, 0, 0, 0, chicago)},
		{chicago, "2026-10-20 09:00", time.Date(2026, time.October, 20, 9, 0, 0, 0, chicago)},
	}
	for _, tc := range cases {
		parser := periodTestParser(now, years.WithLocation(tc.loc))
		got, err := parser.JustParse(tc.value)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got.String()).To(be.Eq(tc.want.String()), "%q in %s", tc.value, tc.loc)
	}
}

func TestWithLocation_ExplicitOffsetUnchanged(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	chisinau := mustLoadLocation(t, "Europe/Chisinau")
	now := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	plain := periodTestParser(now)
	located := periodTestParser(now, years.WithLocation(chisinau))

	for _, value := range []string{"2026-10-20T09:00:00Z", "2026-10-20T09:00:00+05:00"} {
		want, err := plain.JustParse(value)
		be.Require(t, err).To(be.Succeed())
		got, err := located.JustParse(value)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got.Equal(want)).To(be.Eq(true), "%q: got %s, want %s", value, got, want)
	}
}

func TestWithLocation_Periods(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	chisinau := mustLoadLocation(t, "Europe/Chisinau")
	chicago := mustLoadLocation(t, "America/Chicago")
	in := func(loc *time.Location, m time.Month, d int) time.Time {
		return time.Date(2026, m, d, 0, 0, 0, 0, loc)
	}

	cases := []struct {
		name  string
		now   time.Time
		loc   *time.Location
		value string
		want  years.Period
	}{
		// 22:30 UTC on the 20th is already 01:30 on the 21st in Chisinau.
		{
			"today follows the local date",
			time.Date(2026, time.October, 20, 22, 30, 0, 0, time.UTC),
			chisinau,
			"today",
			years.Period{Start: in(chisinau, time.October, 21), End: in(chisinau, time.October, 22)},
		},
		{
			"today in Chicago", time.Date(2026, time.October, 20, 22, 30, 0, 0, time.UTC), chicago, "today",
			years.Period{Start: in(chicago, time.October, 20), End: in(chicago, time.October, 21)},
		},
		{
			"last-week in Chisinau", time.Date(2026, time.October, 20, 22, 30, 0, 0, time.UTC), chisinau, "last-week",
			years.Period{Start: in(chisinau, time.October, 11), End: in(chisinau, time.October, 18)},
		},
		// Chisinau leaves summer time at 03:00 on 2026-10-25, so that day is 25h long.
		{
			"today on the DST day", time.Date(2026, time.October, 25, 10, 0, 0, 0, time.UTC), chisinau, "today",
			years.Period{Start: in(chisinau, time.October, 25), End: in(chisinau, time.October, 26)},
		},
		{
			"plain date on the DST day",
			time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC),
			chisinau,
			"2026-10-25",
			years.Period{Start: in(chisinau, time.October, 25), End: in(chisinau, time.October, 26)},
		},
		{
			"last-week across the DST change",
			time.Date(2026, time.November, 4, 10, 0, 0, 0, time.UTC),
			chisinau,
			"last-week",
			years.Period{Start: in(chisinau, time.October, 25), End: in(chisinau, time.November, 1)},
		},
		{
			"relative offset on the local calendar",
			time.Date(2026, time.October, 28, 10, 0, 0, 0, time.UTC),
			chisinau,
			"7d",
			years.Period{
				Start: time.Date(2026, time.October, 21, 12, 0, 0, 0, chisinau),
				End:   time.Date(2026, time.October, 28, 12, 0, 0, 0, chisinau),
			},
		},
	}
	for _, tc := range cases {
		parser := periodTestParser(tc.now, years.WithLocation(tc.loc))
		got, err := parser.JustParsePeriod(tc.value)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got.Start.String()).To(be.Eq(tc.want.Start.String()), "start: %s", tc.name)
		be.Expect(t, got.End.String()).To(be.Eq(tc.want.End.String()), "end: %s", tc.name)
	}

	dst, err := periodTestParser(
		time.Date(2026, time.October, 25, 10, 0, 0, 0, time.UTC),
		years.WithLocation(chisinau),
	).
		JustParsePeriod("today")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, dst.Duration()).To(be.Eq(25 * time.Hour))
}

func TestWithLocation_UnsetKeepsSplitDefault(t *testing.T) {
	t.Cleanup(years.ResetParserDefaults)

	chicago := mustLoadLocation(t, "America/Chicago")
	parser := periodTestParser(time.Date(2026, time.October, 20, 17, 30, 0, 0, chicago))

	date, err := parser.JustParse("2026-10-20")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, date.String()).To(be.Eq(time.Date(2026, time.October, 20, 0, 0, 0, 0, time.UTC).String()))

	today, err := parser.JustParse("today")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, today.String()).To(be.Eq(time.Date(2026, time.October, 20, 0, 0, 0, 0, chicago).String()))

	nilLoc := periodTestParser(time.Date(2026, time.October, 20, 17, 30, 0, 0, chicago), years.WithLocation(nil))
	date, err = nilLoc.JustParse("2026-10-20")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, date.Location()).To(be.Eq(time.UTC))
}

func TestWithLocation_GlobalDefaults(t *testing.T) {
	chisinau := mustLoadLocation(t, "Europe/Chisinau")
	years.SetStdClock(&StaticClock{now: time.Date(2026, time.October, 20, 22, 30, 0, 0, time.UTC)})
	t.Cleanup(func() {
		years.SetStdClock(&years.StdClock{})
		years.ResetParserDefaults()
	})

	years.ExtendParserDefaults(years.WithLocation(chisinau))

	got, err := years.JustParse("2026-10-20")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, got.String()).To(be.Eq(time.Date(2026, time.October, 20, 0, 0, 0, 0, chisinau).String()))

	period, err := years.JustParsePeriod("today")
	be.Require(t, err).To(be.Succeed())
	be.Expect(t, period.Start.String()).To(be.Eq(time.Date(2026, time.October, 21, 0, 0, 0, 0, chisinau).String()))
}
