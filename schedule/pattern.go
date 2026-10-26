package schedule

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

const (
	daysInWeek = 7

	minutesPerHour = 60
	minutesPerDay  = 24 * minutesPerHour

	// maxSpanEndMinute caps a DaySlot end for pattern purposes: a span starts
	// within its own day (< 24h) and must end before the second midnight
	// (< 48h), so checking a timestamp against its own date and the previous
	// date is always sufficient.
	maxSpanEndMinute = 2 * minutesPerDay
)

// allWeekdays is the full week, used when Days is empty ("every day").
//
//nolint:gochecknoglobals // read-only weekday set
var allWeekdays = []time.Weekday{
	time.Sunday, time.Monday, time.Tuesday, time.Wednesday,
	time.Thursday, time.Friday, time.Saturday,
}

// WeekPattern is a recurring weekly wall-clock pattern: a conjunction of
// "weekday ∈ Days" and "time-of-day ∈ one of Spans".
//   - Days empty  => every day
//   - Spans empty => the whole day
//
// Spans reuse DaySlot; End.Hour may exceed 24 (cross-midnight). A
// cross-midnight span is attributed to the day it starts on:
// {Days: [Sunday], Spans: [{22:00, 29:00}]} covers Sunday 22:00 through
// Monday 05:00.
//
// Unlike Schedule, a WeekPattern has no anchor in absolute time and no
// timezone opinion: Contains reads the wall clock of the given time value
// in its own location. Callers own locality.
type WeekPattern struct {
	Days  []time.Weekday
	Spans []DaySlot
}

// WeekPatterns is a union (OR) of patterns.
type WeekPatterns []WeekPattern

// Contains reports whether t's wall clock matches the pattern.
//
// The attribution rule: Contains(t) is true iff there exists a day d (t's own
// date or the previous date) such that weekday(d) ∈ Days (or Days is empty)
// and t ∈ [d+span.Start, d+span.End) for some span. A cross-midnight span
// belongs to the day it starts on, so a "Sunday nights" pattern includes
// Monday 01:00, matching human usage.
func (p WeekPattern) Contains(t time.Time) bool {
	minuteOfDay := t.Hour()*minutesPerHour + t.Minute()

	// Only t's own date and the previous one can own t (see maxSpanEndMinute).
	for daysAgo := range 2 {
		day := time.Weekday((int(t.Weekday()) + daysInWeek - daysAgo) % daysInWeek)
		if len(p.Days) > 0 && !slices.Contains(p.Days, day) {
			continue
		}
		if p.spansContain(minuteOfDay + daysAgo*minutesPerDay) {
			return true
		}
	}

	return false
}

// Contains reports whether t matches any of the patterns (union semantics).
func (ps WeekPatterns) Contains(t time.Time) bool {
	for _, p := range ps {
		if p.Contains(t) {
			return true
		}
	}
	return false
}

// Validate checks the pattern for sane weekdays and spans: minutes within
// 0-59, spans starting within their own day (< 24h), ending after they start
// and before the second midnight (< 48h).
func (p WeekPattern) Validate() error {
	for _, d := range p.Days {
		if d < time.Sunday || d > time.Saturday {
			return fmt.Errorf("invalid weekday %d", int(d))
		}
	}
	for i, span := range p.Spans {
		if err := validateSpan(span); err != nil {
			return fmt.Errorf("span #%d: %w", i, err)
		}
	}
	return nil
}

// Validate checks every pattern of the union.
func (ps WeekPatterns) Validate() error {
	for i, p := range ps {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("pattern #%d: %w", i, err)
		}
	}
	return nil
}

// MultiSlotSchedule converts the pattern into an equivalent MultiSlotSchedule
// pinned to the given location, so the existing slot math (SlotsForDay,
// PrevMatchingDay, ...) works on patterns. Empty Days/Spans are expanded into
// their explicit meanings (every day / the whole day).
func (p WeekPattern) MultiSlotSchedule(loc *time.Location) MultiSlotSchedule {
	days := slices.Clone(p.Days)
	if len(days) == 0 {
		days = slices.Clone(allWeekdays)
	}

	daySlots := slices.Clone(p.Spans)
	if len(daySlots) == 0 {
		daySlots = []DaySlot{{Start: TimeOfDay{Hour: 0, Minute: 0}, End: TimeOfDay{Hour: 24, Minute: 0}}}
	}

	return MultiSlotSchedule{Days: days, DaySlots: daySlots, Location: loc}
}

// spansContain reports whether minute m, counted from the attribution day's
// midnight (so possibly >= 1440 when attributed to the previous date), falls
// into one of the spans. Empty Spans mean the whole day.
func (p WeekPattern) spansContain(m int) bool {
	if len(p.Spans) == 0 {
		return m < minutesPerDay
	}
	for _, span := range p.Spans {
		if m >= span.Start.totalMinutes() && m < span.End.totalMinutes() {
			return true
		}
	}
	return false
}

// totalMinutes is the time-of-day as minutes from midnight of its day
// (>= 1440 for cross-midnight hours).
func (t TimeOfDay) totalMinutes() int {
	return t.Hour*minutesPerHour + t.Minute
}

// validateSpan checks a single DaySlot used as a pattern/partition span.
func validateSpan(s DaySlot) error {
	if s.Start.Minute < 0 || s.Start.Minute >= minutesPerHour || s.End.Minute < 0 || s.End.Minute >= minutesPerHour {
		return errors.New("minutes must be within 0-59")
	}
	if s.Start.Hour < 0 || s.Start.totalMinutes() >= minutesPerDay {
		return errors.New("span must start within its own day (00:00-23:59)")
	}
	if s.End.totalMinutes() <= s.Start.totalMinutes() {
		return errors.New("span must end after it starts")
	}
	if s.End.totalMinutes() > maxSpanEndMinute {
		return errors.New("span must end within the next day (End <= 48:00)")
	}
	return nil
}
