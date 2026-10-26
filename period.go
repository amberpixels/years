package years

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Period is a half-open time interval [Start, End). Calendar expressions
// denote periods, not instants: "yesterday" is a whole day, "last-week" a
// whole week, "2024-03-06" a day. Start and End serve directly as a range
// query's inclusive lower and exclusive upper bound. A point in time (an
// RFC3339 timestamp, a unix epoch) is an instant period: End equals Start.
type Period struct {
	Start time.Time
	End   time.Time
}

// InstantPeriod wraps a single point in time as a zero-length Period.
func InstantPeriod(t time.Time) Period { return Period{Start: t, End: t} }

func (p Period) IsZero() bool { return p.Start.IsZero() && p.End.IsZero() }

// IsInstant reports whether the period is a zero-length point in time.
func (p Period) IsInstant() bool { return p.End.Equal(p.Start) }

func (p Period) Duration() time.Duration { return p.End.Sub(p.Start) }

// Contains reports whether t falls inside the half-open interval; an instant
// period contains exactly its own point.
func (p Period) Contains(t time.Time) bool {
	if p.IsInstant() {
		return t.Equal(p.Start)
	}
	return !t.Before(p.Start) && t.Before(p.End)
}

// aliasPeriodSpans advances each core alias's period start to its exclusive
// end. Weekend aliases are not here: periodForAlias builds them directly.
//
//nolint:gochecknoglobals // read-only span table, mirrors coreAliasesFor
var aliasPeriodSpans = map[string]func(start time.Time) time.Time{
	"today":     nextDay,
	"yesterday": nextDay,
	"tomorrow":  nextDay,

	aliasLastWorkday: nextDay,

	"this-week": nextWeek,
	"last-week": nextWeek,
	"next-week": nextWeek,

	"this-month": nextMonth,
	"last-month": nextMonth,
	"next-month": nextMonth,

	"this-year": nextYear,
	"last-year": nextYear,
	"next-year": nextYear,
}

// The spans step by calendar date from a period start, so an end on a day
// whose midnight falls in a DST gap is that day's first instant.
func nextDay(t time.Time) time.Time  { return dayStartFrom(t, 1) }
func nextWeek(t time.Time) time.Time { return dayStartFrom(t, daysInWeek) }
func nextMonth(t time.Time) time.Time {
	return startOfDay(t.Year(), t.Month()+1, t.Day(), t.Location())
}

func nextYear(t time.Time) time.Time {
	return startOfDay(t.Year()+1, t.Month(), t.Day(), t.Location())
}

const daysInWeekend = 2

// periodForAlias resolves a registered alias to its full period at base time.
// The start comes from the alias callback so custom overrides are honored; a
// custom alias with no known span is an instant period. Weekends are built as
// [Saturday 00:00, +2d), since the last-weekend instant is the Friday before.
func (p *Parser) periodForAlias(alias string, base time.Time) (Period, bool) {
	cb, known := p.aliases[alias]
	if !known {
		return Period{}, false
	}

	switch alias {
	case aliasNextWeekend, aliasLastWeekend:
		offset := daysUntilSaturday(base)
		if alias == aliasLastWeekend {
			offset = -daysSinceSaturday(base)
		}
		start := dayStartFrom(base, offset)
		return Period{Start: start, End: dayStartFrom(start, daysInWeekend)}, true
	}

	start := cb(base)
	span, ok := aliasPeriodSpans[alias]
	if !ok {
		return InstantPeriod(start), true
	}
	return Period{Start: start, End: span(start)}, true
}

// relativeOffsetRe matches a count and a calendar unit: "12h", "7d", "2w",
// "3mo", "1y". An offset denotes the last N units: [now - N units, now).
//
//nolint:gochecknoglobals // compiled once
var relativeOffsetRe = regexp.MustCompile(`^(\d{1,4})(h|d|w|mo|y)$`)

func parseRelativeOffset(value string, now time.Time) (Period, bool) {
	m := relativeOffsetRe.FindStringSubmatch(value)
	if m == nil {
		return Period{}, false
	}
	n, _ := strconv.Atoi(m[1])
	var start time.Time
	switch m[2] {
	case "h":
		start = now.Add(-time.Duration(n) * time.Hour)
	case "d":
		start = now.AddDate(0, 0, -n)
	case "w":
		start = now.AddDate(0, 0, -n*daysInWeek)
	case "mo":
		start = now.AddDate(0, -n, 0)
	case "y":
		start = now.AddDate(-n, 0, 0)
	}
	return Period{Start: start, End: now}, true
}

// layoutIsDateOnly reports whether a Go layout carries no clock components, so
// its finest unit is a whole day. It looks for Go's hour tokens "15", "03" and
// "3:", none of which appear in date layouts such as "2006-01-02".
func layoutIsDateOnly(layout string) bool {
	return !strings.Contains(layout, "15") &&
		!strings.Contains(layout, "03") &&
		!strings.Contains(layout, "3:")
}

// ParsePeriod parses value into the Period it denotes, using the given layout
// (or all of the parser's accepted layouts when layout is empty):
//   - a registered alias resolves to its calendar period ("yesterday" to that
//     whole day, "last-week" to that whole week);
//   - a relative offset ("7d", "2w") resolves to [now - offset, now);
//   - a value matching a date-only layout resolves to that whole day;
//   - anything else Parse understands (timestamps, epochs) resolves to an
//     instant period.
func (p *Parser) ParsePeriod(layout, value string) (Period, error) {
	if p.acceptAliases {
		if period, ok := p.periodForAlias(value, p.now()); ok {
			return period, nil
		}
	}
	if p.acceptRelativeOffsets {
		if period, ok := parseRelativeOffset(value, p.now()); ok {
			return period, nil
		}
	}

	layouts := p.layouts
	if layout != "" {
		layouts = []string{layout}
	}
	for _, l := range layouts {
		if details := ParseLayout(l); details == nil || details.Format != LayoutFormatGo {
			continue
		}
		t, err := p.parseGoLayout(l, value)
		if err != nil {
			continue
		}
		if layoutIsDateOnly(l) {
			return Period{Start: t, End: nextDay(t)}, nil
		}
		return InstantPeriod(t), nil
	}

	// Everything else Parse understands (epoch forms, timestamp layouts) is a
	// point in time.
	t, err := p.Parse(layout, value)
	if err != nil {
		return Period{}, err
	}
	return InstantPeriod(t), nil
}

// JustParsePeriod is a shortcut for ParsePeriod("", value).
func (p *Parser) JustParsePeriod(value string) (Period, error) {
	return p.ParsePeriod("", value)
}

// ParsePeriod calls ParsePeriod of a default parser.
func ParsePeriod(layout, value string) (Period, error) {
	return DefaultParser().ParsePeriod(layout, value)
}

// JustParsePeriod calls JustParsePeriod of a default parser.
func JustParsePeriod(value string) (Period, error) {
	return DefaultParser().JustParsePeriod(value)
}
