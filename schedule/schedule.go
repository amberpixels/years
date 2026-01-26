package schedule

import (
	"slices"
	"time"
)

// TimeRange represents a time range within a day (hour-based).
type TimeRange struct {
	StartHour int // e.g., 12
	EndHour   int // e.g., 13
}

// Schedule defines working hours for a set of days.
type Schedule struct {
	Days      []time.Weekday // e.g., [Mon, Tue, Wed, Thu, Fri]
	StartHour int            // e.g., 9
	EndHour   int            // e.g., 17
	Gaps      []TimeRange    // optional gaps (e.g., lunch: [{12, 13}])
	Location  *time.Location // timezone
}

// DefaultWorkingHoursSchedule is a standard Monday-Friday 9-17 working hours schedule.
var DefaultWorkingHoursSchedule = Schedule{
	Days:      []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
	StartHour: 9,
	EndHour:   17,
}

// MatchesDay returns true if t's weekday is one of the schedule's days.
func (s Schedule) MatchesDay(t time.Time) bool {
	if s.Location != nil {
		t = t.In(s.Location)
	}

	return slices.Contains(s.Days, t.Weekday())
}

// MatchesTime returns true if t's time-of-day is within schedule hours (excluding gaps).
func (s Schedule) MatchesTime(t time.Time) bool {
	if s.Location != nil {
		t = t.In(s.Location)
	}

	hour := t.Hour()

	// Check if within overall schedule bounds
	if hour < s.StartHour || hour >= s.EndHour {
		return false
	}

	// Check if within any gap
	for _, gap := range s.Gaps {
		if hour >= gap.StartHour && hour < gap.EndHour {
			return false
		}
	}

	return true
}

// Contains returns true if t matches both the day and time of the schedule.
func (s Schedule) Contains(t time.Time) bool {
	return s.MatchesDay(t) && s.MatchesTime(t)
}

// PrevMatchingDay finds the previous day matching the schedule.
// e.g., called on Sunday -> returns Friday (for Mon-Fri schedule).
func (s Schedule) PrevMatchingDay(from time.Time) time.Time {
	if s.Location != nil {
		from = from.In(s.Location)
	}

	// Start from the previous day
	current := from.AddDate(0, 0, -1)

	// Search backwards for up to 7 days (one full week)
	for range 7 {
		if s.MatchesDay(current) {
			return current
		}
		current = current.AddDate(0, 0, -1)
	}

	// If no matching day found in the past week, return the input
	return from
}

// NextMatchingDay finds the next day matching the schedule.
func (s Schedule) NextMatchingDay(from time.Time) time.Time {
	if s.Location != nil {
		from = from.In(s.Location)
	}

	// Start from the next day
	current := from.AddDate(0, 0, 1)

	// Search forwards for up to 7 days (one full week)
	for range 7 {
		if s.MatchesDay(current) {
			return current
		}
		current = current.AddDate(0, 0, 1)
	}

	// If no matching day found in the next week, return the input
	return from
}

// PrevNonMatchingDay finds the previous day not matching the schedule.
// e.g., called on Monday -> returns Sunday (for Mon-Fri schedule).
func (s Schedule) PrevNonMatchingDay(from time.Time) time.Time {
	if s.Location != nil {
		from = from.In(s.Location)
	}

	// Start from the previous day
	current := from.AddDate(0, 0, -1)

	// Search backwards for up to 7 days (one full week)
	for range 7 {
		if !s.MatchesDay(current) {
			return current
		}
		current = current.AddDate(0, 0, -1)
	}

	// If no non-matching day found in the past week, return the input
	return from
}

// NextNonMatchingDay finds the next day not matching the schedule.
func (s Schedule) NextNonMatchingDay(from time.Time) time.Time {
	if s.Location != nil {
		from = from.In(s.Location)
	}

	// Start from the next day
	current := from.AddDate(0, 0, 1)

	// Search forwards for up to 7 days (one full week)
	for range 7 {
		if !s.MatchesDay(current) {
			return current
		}
		current = current.AddDate(0, 0, 1)
	}

	// If no non-matching day found in the next week, return the input
	return from
}
