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

// SlotsForDay generates hour slots for a single day
// If gaps are configured, returns multiple slots (split by gaps).
func (s Schedule) SlotsForDay(day time.Time) []TimeSlot {
	if s.Location != nil {
		day = day.In(s.Location)
	}

	// Check if this day matches the schedule
	if !s.MatchesDay(day) {
		return nil
	}

	// Normalize to start of day
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())

	// If no gaps, return a single slot
	if len(s.Gaps) == 0 {
		return []TimeSlot{
			{
				Start: dayStart.Add(time.Duration(s.StartHour) * time.Hour),
				End:   dayStart.Add(time.Duration(s.EndHour) * time.Hour),
			},
		}
	}

	// Build slots with gaps
	var slots []TimeSlot
	currentStart := s.StartHour

	// Sort gaps by start hour (assuming they're already sorted, but being defensive)
	for _, gap := range s.Gaps {
		// Skip gaps outside schedule bounds
		if gap.EndHour <= s.StartHour || gap.StartHour >= s.EndHour {
			continue
		}

		// Add slot before gap
		gapStart := max(gap.StartHour, s.StartHour)

		if currentStart < gapStart {
			slots = append(slots, TimeSlot{
				Start: dayStart.Add(time.Duration(currentStart) * time.Hour),
				End:   dayStart.Add(time.Duration(gapStart) * time.Hour),
			})
		}

		// Move current start to after gap
		gapEnd := min(gap.EndHour, s.EndHour)
		currentStart = gapEnd
	}

	// Add final slot after last gap
	if currentStart < s.EndHour {
		slots = append(slots, TimeSlot{
			Start: dayStart.Add(time.Duration(currentStart) * time.Hour),
			End:   dayStart.Add(time.Duration(s.EndHour) * time.Hour),
		})
	}

	return slots
}

// SlotsForRange generates slots across multiple days.
func (s Schedule) SlotsForRange(from, to time.Time) []TimeSlot {
	if s.Location != nil {
		from = from.In(s.Location)
		to = to.In(s.Location)
	}

	var slots []TimeSlot

	// Normalize from to start of day
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
