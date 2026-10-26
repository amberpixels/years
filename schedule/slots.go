package schedule

import "time"

// TimeSlot represents a continuous time range.
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// Duration returns the duration of the slot.
func (ts TimeSlot) Duration() time.Duration {
	return ts.End.Sub(ts.Start)
}

// SlotsForDay generates the schedule's hour slots for a single day, split
// around Gaps, which must be sorted by start hour.
func (s Schedule) SlotsForDay(day time.Time) []TimeSlot {
	if s.Location != nil {
		day = day.In(s.Location)
	}

	if !s.MatchesDay(day) {
		return nil
	}

	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())

	if len(s.Gaps) == 0 {
		return []TimeSlot{
			{
				Start: dayStart.Add(time.Duration(s.StartHour) * time.Hour),
				End:   dayStart.Add(time.Duration(s.EndHour) * time.Hour),
			},
		}
	}

	var slots []TimeSlot
	currentStart := s.StartHour

	// Walked in the given order; nothing sorts the gaps here.
	for _, gap := range s.Gaps {
		if gap.EndHour <= s.StartHour || gap.StartHour >= s.EndHour {
			continue
		}

		gapStart := max(gap.StartHour, s.StartHour)

		if currentStart < gapStart {
			slots = append(slots, TimeSlot{
				Start: dayStart.Add(time.Duration(currentStart) * time.Hour),
				End:   dayStart.Add(time.Duration(gapStart) * time.Hour),
			})
		}

		gapEnd := min(gap.EndHour, s.EndHour)
		currentStart = gapEnd
	}

	if currentStart < s.EndHour {
		slots = append(slots, TimeSlot{
			Start: dayStart.Add(time.Duration(currentStart) * time.Hour),
			End:   dayStart.Add(time.Duration(s.EndHour) * time.Hour),
		})
	}

	return slots
}

// SlotsForRange generates slots for every day from from's date through to's
// date, both inclusive.
func (s Schedule) SlotsForRange(from, to time.Time) []TimeSlot {
	if s.Location != nil {
		from = from.In(s.Location)
		to = to.In(s.Location)
	}

	var slots []TimeSlot

	current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	endDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())

	for current.Before(endDate) || current.Equal(endDate) {
		daySlots := s.SlotsForDay(current)
		slots = append(slots, daySlots...)
		current = current.AddDate(0, 0, 1)
	}

	return slots
}

// AvailableMinutes returns total available minutes in schedule for a day.
func (s Schedule) AvailableMinutes(day time.Time) int {
	slots := s.SlotsForDay(day)
	totalMinutes := 0

	for _, slot := range slots {
		totalMinutes += int(slot.Duration().Minutes())
	}

	return totalMinutes
}
