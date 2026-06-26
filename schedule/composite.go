package schedule

import "time"

// CompositeSchedule combines multiple DaySchedule groups into one.
// A day matches if any group matches; slots are the union of all matching groups.
type CompositeSchedule struct {
	Groups   []DaySchedule
	Location *time.Location
}

func (cs CompositeSchedule) MatchesDay(t time.Time) bool {
	if cs.Location != nil {
		t = t.In(cs.Location)
	}
	for _, g := range cs.Groups {
		if g.MatchesDay(t) {
			return true
		}
	}
	return false
}

func (cs CompositeSchedule) SlotsForDay(day time.Time) []TimeSlot {
	if cs.Location != nil {
		day = day.In(cs.Location)
	}
	var slots []TimeSlot
	for _, g := range cs.Groups {
		slots = append(slots, g.SlotsForDay(day)...)
	}
	return slots
}

func (cs CompositeSchedule) PrevMatchingDay(from time.Time) time.Time {
	if cs.Location != nil {
		from = from.In(cs.Location)
	}
	current := from.AddDate(0, 0, -1)
	for range 7 {
		if cs.MatchesDay(current) {
			return current
		}
		current = current.AddDate(0, 0, -1)
	}
	return from
}
