package schedule

import (
	"errors"
	"fmt"
	"time"
)

// DayPart is one named slice of the day within a DayPartition.
type DayPart struct {
	Name string  // canonical singular: "morning", "night", ...
	Span DaySlot // may cross midnight (End.Hour >= 24)
}

// DayPartition is an ordered set of DayParts that must cover the full 24h
// exactly once (no gaps, no overlaps, modulo midnight wrap), so one table
// drives both filtering ("mornings") and classification (PartOf). Validate
// enforces it, which makes PartOf total and a safe group-by key.
type DayPartition []DayPart

// DefaultDayPartition is the canonical partition of the day. Replace it (or
// build a custom Vocabulary) for app-specific slicing, and Validate any custom
// partition first.
//
//nolint:gochecknoglobals // exported, replaceable default
var DefaultDayPartition = DayPartition{
	{Name: "night", Span: DaySlot{Start: TimeOfDay{Hour: 22}, End: TimeOfDay{Hour: 29}}}, // 22:00-05:00, wraps
	{Name: "morning", Span: DaySlot{Start: TimeOfDay{Hour: 5}, End: TimeOfDay{Hour: 12}}},
	{Name: "lunchtime", Span: DaySlot{Start: TimeOfDay{Hour: 12}, End: TimeOfDay{Hour: 14}}},
	{Name: "afternoon", Span: DaySlot{Start: TimeOfDay{Hour: 14}, End: TimeOfDay{Hour: 17}}},
	{Name: "evening", Span: DaySlot{Start: TimeOfDay{Hour: 17}, End: TimeOfDay{Hour: 22}}},
}

// Validate enforces the partition invariant: every minute of the day is
// covered by exactly one part. Overlaps and gaps are reported with the exact
// wall-clock minute so a misconfigured custom partition fails loudly instead
// of silently misclassifying.
func (dp DayPartition) Validate() error {
	if len(dp) == 0 {
		return errors.New("day partition is empty")
	}

	owner := make([]string, minutesPerDay) // minute of day -> owning part name
	for _, part := range dp {
		if part.Name == "" {
			return errors.New("day part has an empty name")
		}
		if err := validateSpan(part.Span); err != nil {
			return fmt.Errorf("part %q: %w", part.Name, err)
		}

		start, end := part.Span.Start.totalMinutes(), part.Span.End.totalMinutes()
		if end-start > minutesPerDay {
			return fmt.Errorf("part %q is longer than 24h", part.Name)
		}
		for m := start; m < end; m++ {
			wallMinute := m % minutesPerDay
			if other := owner[wallMinute]; other != "" {
				return fmt.Errorf("parts %q and %q overlap at %s", other, part.Name, formatMinute(wallMinute))
			}
			owner[wallMinute] = part.Name
		}
	}

	for m, name := range owner {
		if name == "" {
			return fmt.Errorf("no part covers %s", formatMinute(m))
		}
	}

	return nil
}

// PartOf classifies t's wall clock into its part of the day. On a Validate'd
// partition it is total: every minute belongs to exactly one part. A
// partition violating the invariant may return the zero DayPart for
// uncovered minutes.
func (dp DayPartition) PartOf(t time.Time) DayPart {
	minuteOfDay := t.Hour()*minutesPerHour + t.Minute()
	for _, part := range dp {
		start, end := part.Span.Start.totalMinutes(), part.Span.End.totalMinutes()
		// A wrapping part (End >= 24:00) owns early minutes of the next day,
		// so t may fall into it either directly or shifted by one day.
		if (minuteOfDay >= start && minuteOfDay < end) ||
			(minuteOfDay+minutesPerDay >= start && minuteOfDay+minutesPerDay < end) {
			return part
		}
	}
	return DayPart{}
}

// formatMinute renders a minute-of-day as a wall-clock "HH:MM" string.
func formatMinute(m int) string {
	return fmt.Sprintf("%02d:%02d", m/minutesPerHour, m%minutesPerHour)
}
